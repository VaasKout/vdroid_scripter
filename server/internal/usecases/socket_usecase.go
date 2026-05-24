package usecases

import (
	"context"
	"fmt"
	"net"
	"time"
)

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
		if err == nil {
			resultCh <- conn
		}
	}()

	select {
	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("connection timeout: %s", timeoutCtx.Err().Error())
	case conn := <-resultCh:
		return conn, nil
	}
}
