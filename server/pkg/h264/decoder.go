// Package h264 ...
package h264

/*
#cgo pkg-config: libavcodec libavutil
#include <libavcodec/avcodec.h>
#include <string.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"sync"
	"unsafe"
)

// Decoder ...
type Decoder struct {
	mu      sync.Mutex
	frameMu sync.Mutex
	ctx     *C.AVCodecContext
	pkt     *C.AVPacket
	recv    *C.AVFrame
	pending *C.AVFrame
	draw    *C.AVFrame
	closed  bool
}

// NewDecoder ...
func NewDecoder(width int, height int) (*Decoder, error) {
	codec := C.avcodec_find_decoder(C.enum_AVCodecID(C.AV_CODEC_ID_H264))
	if codec == nil {
		return nil, errors.New("h264 decoder not found")
	}

	ctx := C.avcodec_alloc_context3(codec)
	if ctx == nil {
		return nil, errors.New("couldn't allocate codec context")
	}
	ctx.flags = ctx.flags | C.AV_CODEC_FLAG_LOW_DELAY
	ctx.width = C.int(width)
	ctx.height = C.int(height)
	ctx.pix_fmt = C.enum_AVPixelFormat(C.AV_PIX_FMT_YUV420P)

	if code := C.avcodec_open2(ctx, codec, nil); code < 0 {
		C.avcodec_free_context(&ctx)
		return nil, avError("opening codec context failed", code)
	}

	decoder := &Decoder{
		ctx:     ctx,
		pkt:     C.av_packet_alloc(),
		recv:    C.av_frame_alloc(),
		pending: C.av_frame_alloc(),
		draw:    C.av_frame_alloc(),
	}
	if decoder.pkt == nil || decoder.recv == nil ||
		decoder.pending == nil || decoder.draw == nil {
		decoder.Close()
		return nil, errors.New("couldn't allocate decoder buffers")
	}
	return decoder, nil
}

// Decode ...
func (d *Decoder) Decode(packet []byte, keyFrame bool, pts int64) error {
	if len(packet) == 0 {
		return errors.New("packet is empty")
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return errors.New("decoder is closed")
	}

	if code := C.av_new_packet(d.pkt, C.int(len(packet))); code < 0 {
		return avError("packet allocation failed", code)
	}
	C.memcpy(unsafe.Pointer(d.pkt.data), unsafe.Pointer(&packet[0]), C.size_t(len(packet)))
	d.pkt.pts = C.int64_t(pts)
	d.pkt.dts = C.int64_t(pts)
	if keyFrame {
		d.pkt.flags = d.pkt.flags | C.AV_PKT_FLAG_KEY
	}

	code := C.avcodec_send_packet(d.ctx, d.pkt)
	C.av_packet_unref(d.pkt)
	if code < 0 {
		return avError("sending packet failed", code)
	}

	d.receiveFrames()
	return nil
}

func (d *Decoder) receiveFrames() {
	for {
		if code := C.avcodec_receive_frame(d.ctx, d.recv); code < 0 {
			return
		}

		d.frameMu.Lock()
		C.av_frame_unref(d.pending)
		C.av_frame_move_ref(d.pending, d.recv)
		d.frameMu.Unlock()
	}
}

// LatestGray ...
func (d *Decoder) LatestGray() *image.Gray {
	d.frameMu.Lock()
	defer d.frameMu.Unlock()

	frame := d.takeDrawFrame()
	if frame == nil {
		return nil
	}

	width := int(frame.width)
	height := int(frame.height)
	img := image.NewGray(image.Rect(0, 0, width, height))
	copyPlane(img.Pix, img.Stride, frame.data[0], int(frame.linesize[0]), width, height)
	return img
}

// LatestYCbCr ...
func (d *Decoder) LatestYCbCr() *image.YCbCr {
	d.frameMu.Lock()
	defer d.frameMu.Unlock()

	frame := d.takeDrawFrame()
	if frame == nil {
		return nil
	}
	if int(frame.format) != int(C.AV_PIX_FMT_YUV420P) &&
		int(frame.format) != int(C.AV_PIX_FMT_YUVJ420P) {
		return nil
	}
	if frame.data[1] == nil || frame.data[2] == nil {
		return nil
	}

	width := int(frame.width)
	height := int(frame.height)
	img := image.NewYCbCr(image.Rect(0, 0, width, height), image.YCbCrSubsampleRatio420)
	copyPlane(img.Y, img.YStride, frame.data[0], int(frame.linesize[0]), width, height)

	chromaWidth := (width + 1) / 2
	chromaHeight := (height + 1) / 2
	copyPlane(img.Cb, img.CStride, frame.data[1], int(frame.linesize[1]), chromaWidth, chromaHeight)
	copyPlane(img.Cr, img.CStride, frame.data[2], int(frame.linesize[2]), chromaWidth, chromaHeight)
	return img
}

func (d *Decoder) takeDrawFrame() *C.AVFrame {
	if d.closed {
		return nil
	}

	if d.pending.width > 0 && d.pending.data[0] != nil {
		C.av_frame_unref(d.draw)
		C.av_frame_move_ref(d.draw, d.pending)
	}

	if d.draw.width == 0 || d.draw.height == 0 {
		return nil
	}
	if d.draw.data[0] == nil || d.draw.linesize[0] <= 0 {
		return nil
	}
	return d.draw
}

// Size ...
func (d *Decoder) Size() (int, int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.ctx == nil {
		return 0, 0
	}
	return int(d.ctx.width), int(d.ctx.height)
}

// Close ...
func (d *Decoder) Close() {
	d.mu.Lock()
	d.frameMu.Lock()
	defer d.frameMu.Unlock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}
	d.closed = true

	if d.draw != nil {
		C.av_frame_free(&d.draw)
	}
	if d.pending != nil {
		C.av_frame_free(&d.pending)
	}
	if d.recv != nil {
		C.av_frame_free(&d.recv)
	}
	if d.pkt != nil {
		C.av_packet_free(&d.pkt)
	}
	if d.ctx != nil {
		C.avcodec_free_context(&d.ctx)
	}
}

func copyPlane(
	dst []byte,
	dstStride int,
	src *C.uint8_t,
	srcStride int,
	width int,
	height int,
) {
	if src == nil || srcStride < width || width <= 0 || height <= 0 {
		return
	}
	source := unsafe.Slice((*byte)(unsafe.Pointer(src)), srcStride*height)
	for row := range height {
		copy(dst[row*dstStride:row*dstStride+width], source[row*srcStride:row*srcStride+width])
	}
}

func avError(message string, code C.int) error {
	var buf [64]C.char
	C.av_strerror(code, &buf[0], C.size_t(len(buf)))
	return fmt.Errorf("%s: %s", message, C.GoString(&buf[0]))
}
