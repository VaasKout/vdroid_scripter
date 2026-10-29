package scrcpy

import (
	"android_vision_scripter/pkg/models"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"gocv.io/x/gocv"
)

// Connection ...
type Connection interface {
	WriteControlData(serial string, data []byte)
	ReadVideoStream(serial string, clientCh chan []byte)
	GetMatFromLastFrame(serial string, rgb bool) (*gocv.Mat, error)
	GetScreenSize(serial string) (int, int, error)
}

func (s *scrcpyImpl) WriteControlData(serial string, data []byte) {
	if len(data) != models.ControlBytesSize {
		return
	}
	if scrcpyData, ok := s.scrcpyCache.Get(serial); ok {
		n, err := scrcpyData.ControlConn.Write(data)
		if err != nil {
			s.logAPI.Error(
				fmt.Sprintf(
					"err writing control bytes to tcp://127.0.0.1:%d: %s",
					scrcpyData.Port,
					err.Error(),
				),
			)
			return
		}
		if n != models.ControlBytesSize {
			s.logAPI.Error(
				fmt.Sprintf("invalid control buffer size: %d", n),
			)
		}
	}
}

func (s *scrcpyImpl) ReadVideoStream(serial string, clientCh chan []byte) {
	defer func() {
		if clientCh != nil {
			close(clientCh)
		}
	}()
	err := s.sendVideoMetaData(serial, clientCh)
	if err != nil {
		s.logAPI.Error(err.Error())
		return
	}
	scrcpyData, ok := s.scrcpyCache.Get(serial)
	if !ok {
		s.logAPI.Error(fmt.Sprintf("scrcpy server for %s is not active", serial))
		return
	}

	var frameBuf = make([]byte, HeaderSize+InitialPacketSize)
	var handlerCh = make(chan []byte, 10)
	defer close(handlerCh)

	go s.handleFramesManually(scrcpyData.Data, handlerCh)
	var streamErr error
	for {
		_, err := io.ReadFull(scrcpyData.VideoConn, frameBuf[:HeaderSize])
		if err != nil {
			streamErr = fmt.Errorf("reading header: %w", err)
			s.logAPI.Error(streamErr.Error())
			break
		}

		packetSize := int(binary.BigEndian.Uint32(frameBuf[8:12]))
		if packetSize <= 0 || packetSize > MaxPacketSize {
			streamErr = fmt.Errorf("bad header, packet size %d", packetSize)
			s.logAPI.Error(streamErr.Error())
			break
		}

		var totalSize = HeaderSize + packetSize
		if totalSize > len(frameBuf) {
			s.logAPI.Info(fmt.Sprintf("video packet of %d bytes, growing the read buffer", packetSize))
			grown := make([]byte, totalSize)
			copy(grown, frameBuf[:HeaderSize])
			frameBuf = grown
		}

		_, err = io.ReadFull(scrcpyData.VideoConn, frameBuf[HeaderSize:totalSize])
		if err != nil {
			streamErr = fmt.Errorf("reading packet of %d bytes: %w", packetSize, err)
			s.logAPI.Error(streamErr.Error())
			break
		}

		var bufCopy = make([]byte, totalSize)
		copy(bufCopy, frameBuf[:totalSize])

		if clientCh != nil {
			clientCh <- bufCopy
		}
		select {
		case handlerCh <- bufCopy:
		default:
		}
	}
	scrcpyData.markStreamEnded(streamErr)
}

func (s *scrcpyImpl) handleFramesManually(data *DecoderData, handlerCh <-chan []byte) {
	for {
		buf, ok := <-handlerCh
		if !ok {
			s.logAPI.Info("closing manual frames handler... 🛑")
			return
		}
		s.handlePacket(data, buf)
	}
}

func (s *scrcpyImpl) GetMatFromLastFrame(serial string, rgb bool) (*gocv.Mat, error) {
	scrcpyData, ok := s.scrcpyCache.Get(serial)
	if !ok || scrcpyData == nil {
		return nil, fmt.Errorf("video connecton for %s is not active", serial)
	}
	if err := scrcpyData.streamEnded(); err != nil {
		return nil, fmt.Errorf("video stream for %s ended: %w", serial, err)
	}

	mat := s.frameToMat(scrcpyData.Data, rgb)
	if mat == nil {
		return nil, nil
	}

	defer mat.Close()
	if mat.Empty() {
		return nil, nil
	}

	newMat := mat.Clone()
	return &newMat, nil
}

func (s *scrcpyImpl) GetScreenSize(serial string) (int, int, error) {
	scrcpyData, ok := s.scrcpyCache.Get(serial)
	if !ok || scrcpyData == nil || scrcpyData.Data == nil {
		return 0, 0, fmt.Errorf("video connecton for %s is not active", serial)
	}

	width, height := scrcpyData.Data.GetSize()
	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("screen size for %s is not available", serial)
	}
	return width, height, nil
}

func (s *scrcpyImpl) sendVideoMetaData(serial string, ch chan []byte) error {
	scrcpyData, ok := s.scrcpyCache.Get(serial)
	if !ok {
		return fmt.Errorf("video connecton for %s is not active", serial)
	}
	buf := make([]byte, 4)
	width, height := scrcpyData.Data.GetSize()
	binary.BigEndian.PutUint32(buf, uint32(width))
	if ch != nil {
		ch <- buf
	}
	buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(height))
	if ch != nil {
		ch <- buf
	}
	return nil
}

func (s *scrcpyImpl) initConnections(serial string, port int) error {
	newVideoConn, err := s.startSocketConn(port)
	if err != nil {
		return err
	}
	s.logAPI.Info(fmt.Sprintf("video connection started: tcp://127.0.0.1:%d ✅", port))

	newControlConn, err := s.startSocketConn(port)
	if err != nil {
		newVideoConn.Close()
		return err
	}
	s.logAPI.Info(fmt.Sprintf("control connection started: tcp://127.0.0.1:%d ✅", port))

	width, height, err := s.readMetaData(newVideoConn)
	if err != nil {
		newControlConn.Close()
		newVideoConn.Close()
		return fmt.Errorf("couldn't receive metadata: %v", err.Error())
	}

	var decoderData = &DecoderData{}
	err = decoderData.Allocate(width, height)
	if err != nil {
		decoderData.Free()
		newControlConn.Close()
		newVideoConn.Close()
		return err
	}

	var scrcpyData = &Data{
		Port:        port,
		Data:        decoderData,
		VideoConn:   newVideoConn,
		ControlConn: newControlConn,
	}

	s.scrcpyCache.Add(serial, scrcpyData)
	return nil
}

func (s *scrcpyImpl) startSocketConn(connPort int) (net.Conn, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", connPort))
	if err != nil {
		s.logAPI.Error(fmt.Sprintf("can't start socket conn: %s", err.Error()))
		return nil, err
	}
	return conn, nil
}

func (s *scrcpyImpl) readMetaData(conn net.Conn) (width, height int, err error) {
	var dummyByte = make([]byte, 1)
	_, err = conn.Read(dummyByte)
	if err != nil {
		return width, height, fmt.Errorf("failed to read dummy byte, err:%v", err.Error())
	}
	s.logAPI.Info(fmt.Sprintf("dummy byte: %v", dummyByte))

	var deviceName = make([]byte, 64)
	_, err = conn.Read(deviceName)
	if err != nil {
		return width, height, err
	}
	s.logAPI.Info(fmt.Sprintf("name: %s", string(deviceName)))

	var codecID = make([]byte, 4)
	_, err = conn.Read(codecID)
	if err != nil {
		return width, height, err
	}
	s.logAPI.Info(fmt.Sprintf("codecID: %v", codecID))

	var widthBuf = make([]byte, 4)
	_, err = conn.Read(widthBuf)
	if err != nil {
		return width, height, err
	}
	width = int(binary.BigEndian.Uint32(widthBuf))
	s.logAPI.Info(fmt.Sprintf("width: %d", width))

	var heightBuf = make([]byte, 4)
	_, err = conn.Read(heightBuf)
	if err != nil {
		return width, height, err
	}
	height = int(binary.BigEndian.Uint32(heightBuf))
	s.logAPI.Info(fmt.Sprintf("height: %d", height))

	if width == 0 || height == 0 {
		return width, height, fmt.Errorf("width: %d, height: %d", width, height)
	}

	return width, height, nil
}
