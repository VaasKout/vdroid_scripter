package scrcpy

import (
	"android_vision_scripter/internal/h264"
	"encoding/binary"
	"fmt"

	"gocv.io/x/gocv"
)

// Size of different buffers
const (
	BufSize    = 1 * 1024 * 1024 //1MB
	HeaderSize = 12
)

// PTS flags
const (
	SCPacketFlagConfig   uint64 = 1 << 63
	SCPacketFlagKeyFrame uint64 = 1 << 62

	SCPacketPtsMask uint64 = SCPacketFlagKeyFrame - 1
)

// DecoderData ...
type DecoderData struct {
	Decoder *h264.Decoder

	Buf            []byte
	HeaderBuf      []byte
	ConfigFrameBuf []byte
}

// Free ...
func (d *DecoderData) Free() {
	if d == nil {
		return
	}
	if d.Decoder != nil {
		d.Decoder.Close()
	}
	d.HeaderBuf = []byte{}
	d.ConfigFrameBuf = []byte{}
	d.Buf = []byte{}
}

// Allocate ...
func (d *DecoderData) Allocate(width, height int) error {
	if d == nil {
		return fmt.Errorf("data is nil")
	}

	decoder, err := h264.NewDecoder(width, height)
	if err != nil {
		return err
	}

	d.Decoder = decoder
	d.HeaderBuf = make([]byte, HeaderSize)
	d.Buf = make([]byte, BufSize)
	return nil
}

// GetSize ...
func (d *DecoderData) GetSize() (width, height int) {
	if d == nil || d.Decoder == nil {
		return 0, 0
	}
	return d.Decoder.Size()
}

func (s *scrcpyImpl) handlePackets(data *DecoderData) error {
	if data == nil || data.Decoder == nil {
		return fmt.Errorf("decoder data is nil")
	}
	if len(data.HeaderBuf) < HeaderSize {
		return fmt.Errorf("data is empty")
	}

	pts := binary.BigEndian.Uint64(data.HeaderBuf[0:8])
	packetSize := binary.BigEndian.Uint32(data.HeaderBuf[8:12])
	if pts <= 0 || int(packetSize) == 0 {
		return nil
	}

	var configSize = len(data.ConfigFrameBuf)
	var neededSpace = configSize + int(packetSize)
	if configSize > 0 {
		copy(data.Buf[configSize:], data.Buf[:len(data.Buf)])
		copy(data.Buf[:configSize], data.ConfigFrameBuf)
		data.ConfigFrameBuf = []byte{}
	}

	if pts&SCPacketFlagConfig != 0 {
		data.ConfigFrameBuf = append([]byte{}, data.Buf[:neededSpace]...)
		return nil
	}

	keyFrame := pts&SCPacketFlagKeyFrame != 0
	err := data.Decoder.Decode(
		data.Buf[:neededSpace],
		keyFrame,
		int64(pts&SCPacketPtsMask),
	)
	if err != nil {
		s.logAPI.Error(fmt.Sprintf("decoding packet failed: %s", err.Error()))
	}
	return nil
}

func (s *scrcpyImpl) frameToMat(data *DecoderData, rgb bool) *gocv.Mat {
	if data == nil || data.Decoder == nil {
		return nil
	}

	if rgb {
		img := data.Decoder.LatestYCbCr()
		if img == nil {
			return nil
		}
		mat, err := gocv.ImageToMatRGB(img)
		if err != nil {
			s.logAPI.Error(err.Error())
			mat.Close()
			return nil
		}
		return &mat
	}

	img := data.Decoder.LatestGray()
	if img == nil {
		return nil
	}
	mat, err := gocv.ImageGrayToMatGray(img)
	if err != nil {
		s.logAPI.Error(err.Error())
		mat.Close()
		return nil
	}
	return &mat
}
