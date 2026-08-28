package usecases

import (
	"android_vision_scripter/internal/scrcpy"
	"android_vision_scripter/pkg/models"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"gocv.io/x/gocv"
)

// ScrcpyUseCase ...
type ScrcpyUseCase interface {
	StartScrcpyServer(serial string, serverPort int) bool
	AcceptVideoConnections(
		ctx context.Context,
		serial string,
	)
	AcceptCvConnection(
		ctx context.Context,
		serial string,
	)
	AcceptControlConnection(
		ctx context.Context,
		serial string,
	)
}

func (i *interactorImpl) StartScrcpyServer(serial string, serverPort int) bool {
	streamURL := i.scrcpy.StartScrcpyServer(serial, serverPort)
	return streamURL != ""
}

func (i *interactorImpl) AcceptVideoConnections(
	ctx context.Context,
	serial string,
) {
	var session *models.Session
	if result, ok := i.sessionsCache.Get(serial); ok {
		session = &result
	}
	if session == nil || session.VideoPort == 0 {
		return
	}

	videoListener, err := i.startSocketListener(session.VideoPort)
	if err != nil {
		var errMsg = fmt.Sprintf(
			"couldn't start video listener on port %d for %s",
			session.VideoPort,
			serial,
		)
		i.logger.Error(errMsg)
		return
	}
	defer videoListener.Close()
	defer i.logger.Info("closing video listener... 🛑")

	videoListenerConn, err := i.acceptListenerWithTimeout(videoListener)
	if err != nil {
		i.logger.Error(fmt.Sprintf("closing video listener with err: %s", err.Error()))
		return
	}
	defer videoListenerConn.Close()
	defer i.logger.Info("closing video connection... 🛑")

	i.logger.Info(
		fmt.Sprintf(
			"start listening video socket on port %d for %s... ⏳",
			session.VideoPort,
			serial,
		),
	)

	i.copyVideoStream(ctx, serial, videoListenerConn)
}

func (i *interactorImpl) copyVideoStream(
	ctx context.Context,
	serial string,
	to net.Conn,
) {
	defer i.logger.Info("closing videostream... 🛑")

	var ch = make(chan []byte, 10)
	go i.scrcpy.ReadVideoStream(serial, ch)

	var clientDisconnected = make(chan struct{}, 1)
	go func() {
		buf := make([]byte, 4)
		n, err := to.Read(buf)
		if err != nil || n == 0 {
			close(clientDisconnected)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-clientDisconnected:
			return
		case buf, ok := <-ch:
			if !ok {
				return
			}
			_, err := to.Write(buf)
			if err != nil {
				return
			}
		}
	}
}

func (i *interactorImpl) AcceptCvConnection(
	ctx context.Context,
	serial string,
) {
	var session *models.Session
	if result, ok := i.sessionsCache.Get(serial); ok {
		session = &result
	}
	if session == nil ||
		session.VideoPort == 0 ||
		session.CVPort == 0 {
		return
	}

	cvListener, err := i.startSocketListener(session.CVPort)
	if err != nil {
		var errMsg = fmt.Sprintf(
			"couldn't start cv listener on port %d for %s",
			session.CVPort,
			serial,
		)
		i.logger.Error(errMsg)
		return
	}
	defer cvListener.Close()
	defer i.logger.Info("closing cv listener... 🛑")

	clientCvConn, err := i.acceptListenerWithTimeout(cvListener)
	if err != nil {
		i.logger.Error(fmt.Sprintf("closing cv listener with err: %s", err.Error()))
		return
	}
	defer clientCvConn.Close()
	defer i.logger.Info("closing client cv connection... 🛑")

	i.logger.Info(
		fmt.Sprintf(
			"start listening cv socket on port %d for %s... ⏳",
			session.CVPort,
			serial,
		),
	)

	var cvModeCh = make(chan int, 1)
	go i.readCVClient(ctx, clientCvConn, cvModeCh)
	i.writeToCVClient(ctx, serial, clientCvConn, cvModeCh)
}

func (i *interactorImpl) readCVClient(
	ctx context.Context,
	cvListenerConn net.Conn,
	cvModeCh chan<- int,
) {
	defer close(cvModeCh)
	cvModeBuf := make([]byte, 4)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := io.ReadFull(cvListenerConn, cvModeBuf)
			if err != nil {
				i.logger.Error(fmt.Sprintf("cv client disconnected: %s", err.Error()))
				return
			}

			cvMode := binary.BigEndian.Uint32(cvModeBuf)
			cvModeCh <- int(cvMode)
		}
	}
}

func (i *interactorImpl) writeToCVClient(
	ctx context.Context,
	serial string,
	cvListenerConn net.Conn,
	cvModeCh <-chan int,
) {
	var cvMode = scrcpy.NoCV
	for {
		select {
		case <-ctx.Done():
			return
		case newMode, ok := <-cvModeCh:
			if !ok {
				return
			}
			cvMode = newMode

			if newMode == scrcpy.NoCV {
				continue
			}

			rects, err := i.getRectangles(serial, cvMode, false)
			if err != nil {
				return
			}

			err = i.sendRectangles(rects, cvListenerConn)
			if err != nil {
				return
			}
		default:
			if cvMode == scrcpy.NoCV {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			rects, err := i.getRectangles(serial, cvMode, true)
			if err != nil {
				return
			}

			err = i.sendRectangles(rects, cvListenerConn)
			if err != nil {
				return
			}

			if len(rects) == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if cvMode == scrcpy.Yolo {
				sleepUntilNextSecond()
			}
		}
	}
}

func (i *interactorImpl) getRectangles(
	serial string,
	cvMode int,
	frameOnly bool,
) ([]models.Rectangle, error) {
	var rgb = cvMode == scrcpy.Yolo
	mat, _ := i.scrcpy.GetMatFromLastFrame(serial, rgb)
	if mat == nil && !frameOnly {
		mat = i.getScreenshotMat(serial, cvMode)
	}
	if mat == nil {
		return []models.Rectangle{}, nil
	}
	defer mat.Close()

	return i.detectRectangles(cvMode, mat)
}

func (i *interactorImpl) getScreenshotMat(
	serial string,
	cvMode int,
) *gocv.Mat {
	screenshot := i.cmd.ScreenShot(serial)
	if screenshot == "" {
		return nil
	}

	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	if img.Empty() {
		img.Close()
		return nil
	}
	if cvMode != scrcpy.CVRects {
		return &img
	}

	defer img.Close()
	gray := gocv.NewMat()
	gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	return &gray
}

func (i *interactorImpl) detectRectangles(
	cvMode int,
	mat *gocv.Mat,
) ([]models.Rectangle, error) {
	if cvMode == scrcpy.Yolo {
		return i.yolo.DetectLabels(*mat), nil
	}
	rects, err := i.cv.FindAllRectangles(mat)
	if err != nil {
		i.logger.Error(err.Error())
		return []models.Rectangle{}, err
	}
	return models.ImgRectanglesToDomain(rects), nil
}

func (i *interactorImpl) sendRectangles(
	rects []models.Rectangle,
	cvListenerConn net.Conn,
) error {
	if len(rects) == 0 {
		return nil
	}
	bytes, err := json.Marshal(rects)
	if err != nil {
		i.logger.Error(err.Error())
		return err
	}

	err = binary.Write(cvListenerConn, binary.BigEndian, uint32(len(bytes)))
	if err != nil {
		i.logger.Error(fmt.Sprintf("cv client disconnected: %s 🛑", err.Error()))
		return err
	}

	_, err = cvListenerConn.Write(bytes)
	return err
}

func (i *interactorImpl) AcceptControlConnection(
	ctx context.Context,
	serial string,
) {
	var session *models.Session
	if result, ok := i.sessionsCache.Get(serial); ok {
		session = &result
	}
	if session == nil || session.ControlPort == 0 {
		return
	}
	controlListener, err := i.startSocketListener(session.ControlPort)
	if err != nil {
		var errMsg = fmt.Sprintf(
			"couldn't start control listener on port %d for %s",
			session.ControlPort,
			serial,
		)
		i.logger.Error(errMsg)
		return
	}
	defer controlListener.Close()
	defer i.logger.Info("closing control listener... 🛑")

	controlClientConn, err := i.acceptListenerWithTimeout(controlListener)
	if err != nil {
		i.logger.Error(fmt.Sprintf("closing control listener with err: %s", err.Error()))
		return
	}
	defer controlClientConn.Close()
	defer i.logger.Info("closing client control connection... 🛑")

	i.logger.Info(
		fmt.Sprintf(
			"start listening control socket on port %d for %s... ⏳",
			session.ControlPort,
			serial,
		),
	)

	i.sendControlDataFromClient(ctx, serial, controlClientConn)
}

func (i *interactorImpl) sendControlDataFromClient(
	ctx context.Context,
	serial string,
	from net.Conn,
) {
	var controlDataBuffer = make([]byte, models.ControlBytesSize)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := io.ReadFull(from, controlDataBuffer)
			if err != nil {
				i.logger.Error(fmt.Sprintf("control client for %s disconnected 🛑", serial))
				return
			}
			i.scrcpy.WriteControlData(serial, controlDataBuffer)
		}
	}
}

func (i *interactorImpl) startSocketListener(
	listenerPort int,
) (net.Listener, error) {
	var connection = fmt.Sprintf("0.0.0.0:%d", listenerPort)
	listener, err := net.Listen("tcp", connection)
	if err != nil {
		i.logger.Error(fmt.Sprintf("can't start socket listener: %s", err.Error()))
		return nil, err
	}
	return listener, nil
}

func (i *interactorImpl) acceptListenerWithTimeout(
	listener net.Listener,
) (net.Conn, error) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resultCh := make(chan net.Conn, 1)
	defer close(resultCh)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		if timeoutCtx.Err() != nil {
			conn.Close()
			return
		}
		resultCh <- conn
	}()

	select {
	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("connection timeout: %s", timeoutCtx.Err().Error())
	case conn := <-resultCh:
		return conn, nil
	}
}
