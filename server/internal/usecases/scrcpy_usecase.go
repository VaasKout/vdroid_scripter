package usecases

import (
	"android_vision_scripter/pkg/models"
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

// ScrcpyUseCase ...
type ScrcpyUseCase interface {
	StartScrcpyServer(serial string, serverPort int) bool
	AcceptVideoConnections(
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
