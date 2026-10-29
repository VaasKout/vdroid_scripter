package scrcpy

import (
	"android_vision_scripter/pkg/h264"
	"encoding/binary"
	"fmt"

	"gocv.io/x/gocv"
)

// Size of different buffers
const (
	HeaderSize        = 12
	InitialPacketSize = 1 * 1024 * 1024
	MaxPacketSize     = 64 * 1024 * 1024
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
	d.ConfigFrameBuf = []byte{}
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
	return nil
}

// GetSize ...
func (d *DecoderData) GetSize() (width, height int) {
	if d == nil || d.Decoder == nil {
		return 0, 0
	}
	return d.Decoder.Size()
}

func (s *scrcpyImpl) handlePacket(data *DecoderData, packet []byte) error {
	if data == nil || data.Decoder == nil {
		return fmt.Errorf("decoder data is nil")
	}
	if len(packet) < HeaderSize {
		return fmt.Errorf("packet is shorter than its header")
	}

	pts := binary.BigEndian.Uint64(packet[0:8])
	packetSize := int(binary.BigEndian.Uint32(packet[8:12]))
	if pts <= 0 || packetSize == 0 || len(packet) < HeaderSize+packetSize {
		return nil
	}
	payload := packet[HeaderSize : HeaderSize+packetSize]

	if pts&SCPacketFlagConfig != 0 {
		data.ConfigFrameBuf = append([]byte{}, payload...)
		return nil
	}

	frame := payload
	if len(data.ConfigFrameBuf) > 0 {
		frame = make([]byte, 0, len(data.ConfigFrameBuf)+len(payload))
		frame = append(frame, data.ConfigFrameBuf...)
		frame = append(frame, payload...)
		data.ConfigFrameBuf = []byte{}
	}

	keyFrame := pts&SCPacketFlagKeyFrame != 0
	err := data.Decoder.Decode(frame, keyFrame, int64(pts&SCPacketPtsMask))
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
