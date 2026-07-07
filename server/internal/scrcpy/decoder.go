package scrcpy

import (
	"encoding/binary"
	"fmt"
	"image"

	"sync"

	"github.com/asticode/go-astiav"
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
	CodecContext *astiav.CodecContext
	Decoder      *astiav.Codec
	Pkt          *astiav.Packet
	Frame        *astiav.Frame
	PendingFrame *astiav.Frame
	DrawFrame    *astiav.Frame

	Buf            []byte
	HeaderBuf      []byte
	ConfigFrameBuf []byte
	GrayImg        *image.Gray
	YcbCrImg       *image.YCbCr

	frameMu sync.Mutex
	pktMu   sync.Mutex
}

// Free ...
func (d *DecoderData) Free() {
	if d == nil {
		return
	}
	d.pktMu.Lock() //prevent segmentation fault in CodecContext.SendPacket
	d.frameMu.Lock()
	d.Frame.Free()
	d.PendingFrame.Free()
	d.DrawFrame.Free()
	d.Pkt.Free()
	d.CodecContext.Free()
	d.HeaderBuf = []byte{}
	d.ConfigFrameBuf = []byte{}
	d.Buf = []byte{}
	d.GrayImg = nil
	d.YcbCrImg = nil
	d.pktMu.Unlock()
	d.frameMu.Unlock()
	d = nil
}

// Allocate ...
func (d *DecoderData) Allocate(width, height int) error {
	if d == nil {
		return fmt.Errorf("data is nil")
	}
	codecID := astiav.CodecIDH264
	d.Decoder = astiav.FindDecoder(codecID)
	d.CodecContext = astiav.AllocCodecContext(d.Decoder)

	var newFlags = d.CodecContext.Flags().Add(astiav.CodecContextFlagLowDelay)
	d.CodecContext.SetFlags(newFlags)
	d.CodecContext.SetWidth(width)
	d.CodecContext.SetHeight(height)
	d.CodecContext.SetPixelFormat(astiav.PixelFormatYuv420P)
	d.Pkt = astiav.AllocPacket()
	d.Frame = astiav.AllocFrame()
	d.PendingFrame = astiav.AllocFrame()
	d.DrawFrame = astiav.AllocFrame()

	d.HeaderBuf = make([]byte, HeaderSize)
	d.Buf = make([]byte, BufSize)
	d.GrayImg = &image.Gray{}
	d.YcbCrImg = &image.YCbCr{}

	// Open codec context
	if err := d.CodecContext.Open(d.Decoder, nil); err != nil {
		return fmt.Errorf("opening codec context failed: %w", err)
	}
	return nil
}

// GetSize ...
func (d *DecoderData) GetSize() (width, height int) {
	if d.CodecContext != nil {
		return d.CodecContext.Width(), d.CodecContext.Height()
	}
	return 0, 0
}

func (s *scrcpyImpl) handlePackets(data *DecoderData) error {
	if data == nil {
		return fmt.Errorf("decoder data is nil")
	}
	data.pktMu.Lock()
	defer data.pktMu.Unlock()

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

	err := data.Pkt.FromData(data.Buf[:neededSpace])
	if err != nil {
		s.logAPI.Error(fmt.Sprintf("data set failed: %s", err.Error()))
		return err
	}
	defer data.Pkt.Unref()

	if pts&SCPacketFlagKeyFrame != 0 {
		var updatedFlags = data.Pkt.Flags().Add(astiav.PacketFlagKey)
		data.Pkt.SetFlags(updatedFlags)
	}

	var initConfigBuf = false
	if pts&SCPacketFlagConfig != 0 {
		data.Pkt.SetPts(astiav.NoPtsValue)
		initConfigBuf = true
	} else {
		data.Pkt.SetPts(int64(pts & SCPacketPtsMask))
	}

	data.Pkt.SetDts(data.Pkt.Pts())

	if initConfigBuf {
		data.ConfigFrameBuf = data.Pkt.Data()
		return nil
	}

	if err := data.CodecContext.SendPacket(data.Pkt); err != nil {
		s.logAPI.Error(fmt.Sprintf("sending packet failed: %v", err))
	}

	receiveFrames(data)
	return nil
}

func receiveFrames(data *DecoderData) {
	for {
		if err := data.CodecContext.ReceiveFrame(data.Frame); err != nil {
			data.Frame.Unref()
			if err == astiav.ErrEagain || err == astiav.ErrEof {
				return
			}
		}

		data.frameMu.Lock()
		data.PendingFrame.Unref()
		data.PendingFrame.MoveRef(data.Frame)
		data.frameMu.Unlock()
	}
}

func (s *scrcpyImpl) frameToMat(data *DecoderData, rgb bool) *gocv.Mat {
	data.frameMu.Lock()
	data.DrawFrame.Unref()
	data.DrawFrame.MoveRef(data.PendingFrame)
	data.frameMu.Unlock()

	var width = data.DrawFrame.Width()
	var height = data.DrawFrame.Height()
	if width == 0 && height == 0 {
		return nil
	}

	var frameData = data.DrawFrame.Data()
	if frameData == nil || data.YcbCrImg == nil {
		return nil
	}

	if rgb {
		// Populate the image with the frame's data
		if err := frameData.ToImage(data.YcbCrImg); err != nil {
			return nil
		}

		// Convert the Go image to a gocv.Mat
		mat, err := gocv.ImageToMatRGB(data.YcbCrImg)
		if err != nil {
			s.logAPI.Error(err.Error())
			mat.Close()
			return nil
		}
		return &mat
	}

	if err := frameData.ToImage(data.GrayImg); err != nil {
		return nil
	}

	mat, err := gocv.ImageGrayToMatGray(data.GrayImg)
	if err != nil {
		s.logAPI.Error(err.Error())
		mat.Close()
		return nil
	}

	return &mat
}
