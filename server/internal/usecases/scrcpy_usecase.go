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

// ClientConnection ...
type ClientConnection struct {
	ServerPort  int
	VideoPort   int
	CVPort      int
	ControlPort int
	DoneCh      chan struct{}
}

// ScrcpyUseCase ...
type ScrcpyUseCase interface {
	StartScrcpyServer(serial string, serverPort int) bool
	CloseConnection(serial string)
	GetPortsJSON(serial string) map[string]string
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

func (i *interactorImpl) StartScrcpyServer(
	serial string,
	basePort int,
) bool {
	i.logger.Info(fmt.Sprintf("closing old connections for %s... ⏳", serial))
	i.CloseConnection(serial)

	clientConnection := i.initPorts(basePort)
	if clientConnection == nil {
		return false
	}
	i.clientsCache.Add(serial, *clientConnection)

	streamURL := i.scrcpy.StartScrcpyServer(serial, clientConnection.ServerPort)
	if streamURL == "" {
		return false
	}

	i.setScrcpyState(serial, true)
	return true
}

func (i *interactorImpl) CloseConnection(serial string) {
	i.logger.Info(
		fmt.Sprintf(
			"closing scrcpy connection for %s... 🛑",
			serial,
		),
	)
	i.scrcpy.CloseScrcpyServer(serial)
	if connection, ok := i.clientsCache.Get(serial); ok {
		close(connection.DoneCh)
		i.clientsCache.Delete(serial)
	}
	i.setScrcpyState(serial, false)
}

func (i *interactorImpl) GetPortsJSON(serial string) map[string]string {
	if result, ok := i.clientsCache.Get(serial); ok {
		return map[string]string{
			"video_port":   fmt.Sprintf("%d", result.VideoPort),
			"cv_port":      fmt.Sprintf("%d", result.CVPort),
			"control_port": fmt.Sprintf("%d", result.ControlPort),
		}
	}
	return map[string]string{}
}

func (i *interactorImpl) initPorts(basePort int) *ClientConnection {
	var videoPort = basePort + 1
	var cvPort = videoPort + 1
	var controlPort = cvPort + 1

	var cacheMap = i.clientsCache.GetMap()
	if len(cacheMap) == 0 {
		return &ClientConnection{
			ServerPort:  basePort,
			VideoPort:   videoPort,
			CVPort:      cvPort,
			ControlPort: controlPort,
			DoneCh:      make(chan struct{}),
		}
	}

	var biggestPort = controlPort
	for _, conn := range cacheMap {
		if conn.ControlPort > biggestPort {
			biggestPort = conn.ControlPort
		}
	}

	serverPort := biggestPort + 1
	videoPort = serverPort + 1
	cvPort = videoPort + 1
	controlPort = cvPort + 1

	return &ClientConnection{
		ServerPort:  serverPort,
		VideoPort:   videoPort,
		CVPort:      cvPort,
		ControlPort: controlPort,
	}
}

func (i *interactorImpl) AcceptVideoConnections(
	ctx context.Context,
	serial string,
) {
	var clientConnection *ClientConnection
	if result, ok := i.clientsCache.Get(serial); ok {
		clientConnection = &result
	}
	if clientConnection == nil || clientConnection.VideoPort == 0 {
		return
	}

	videoListener, err := i.startSocketListener(clientConnection.VideoPort)
	if err != nil {
		var errMsg = fmt.Sprintf(
			"couldn't start video listener on port %d for %s",
			clientConnection.VideoPort,
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
			clientConnection.VideoPort,
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
		case _, _ = <-clientDisconnected:
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
	var clientConnection *ClientConnection
	if result, ok := i.clientsCache.Get(serial); ok {
		clientConnection = &result
	}
	if clientConnection == nil ||
		clientConnection.VideoPort == 0 ||
		clientConnection.CVPort == 0 {
		return
	}

	cvListener, err := i.startSocketListener(clientConnection.CVPort)
	if err != nil {
		var errMsg = fmt.Sprintf(
			"couldn't start cv listener on port %d for %s",
			clientConnection.CVPort,
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
			clientConnection.CVPort,
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
	var clientConnection *ClientConnection
	if result, ok := i.clientsCache.Get(serial); ok {
		clientConnection = &result
	}
	if clientConnection == nil || clientConnection.ControlPort == 0 {
		return
	}
	controlListener, err := i.startSocketListener(clientConnection.ControlPort)
	if err != nil {
		var errMsg = fmt.Sprintf(
			"couldn't start control listener on port %d for %s",
			clientConnection.ControlPort,
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
			clientConnection.ControlPort,
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
