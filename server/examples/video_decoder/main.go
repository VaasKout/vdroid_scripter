// Package main ...
package main

import (
	"android_vision_scripter/internal/h264"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"gocv.io/x/gocv"
)

func init() {
	// macOS (Cocoa/AppKit) requires NSWindow to be created on the main OS thread.
	// Lock the main goroutine to the startup thread so gocv.NewWindow runs there.
	// Linux/X11 has no such restriction, so only do this on macOS.
	if runtime.GOOS == "darwin" {
		runtime.LockOSThread()
	}
}

// Size of different buffers
const (
	BufSize    = 1 * 1024 * 1024 //1MB
	HeaderSize = 12
)

// Scrcpy properties and adb commands
const (
	ScrcpyVersion = "3.3.4"

	ScrcpyLinkFormat = "https://github.com/Genymobile/scrcpy/releases/download/v%s/scrcpy-server-v%s"
	ScrcpyFileFormat = "scrcpy-v%s"
	BasePath         = "android_vision_scripter"
	ScrcpyDir        = "scrcpy"

	PushFile          = "adb -s %s push %s /data/local/tmp/scrcpy-server.jar"
	ForwardTCPPort    = "adb -s %s forward tcp:1234 localabstract:scrcpy"
	StartScrcpyServer = "adb -s %s shell CLASSPATH=/data/local/tmp/scrcpy-server.jar app_process / com.genymobile.scrcpy.Server %s log_level=verbose tunnel_forward=true audio=false control=false cleanup=false"
)

// PTS flags
const (
	SCPacketFlagConfig   uint64 = 1 << 63
	SCPacketFlagKeyFrame uint64 = 1 << 62

	SCPacketPtsMask uint64 = SCPacketFlagKeyFrame - 1
)

// DecoderData ...
type DecoderData struct {
	Decoder        *h264.Decoder
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
	d.Buf = []byte{}
	d.HeaderBuf = []byte{}
	d.ConfigFrameBuf = []byte{}
}

// Allocate ...
func (d *DecoderData) Allocate(width, height int) error {
	if d == nil {
		return errors.New("data is nil")
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

func main() {
	var serial string
	if len(os.Args) > 1 {
		serial = strings.TrimSpace(os.Args[1])
	}
	if serial == "" {
		panic(fmt.Errorf("missing device serial: usage: go run . SERIAL"))
	}

	if err := startScrcpy(serial); err != nil {
		panic(err)
	}
	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatalf("Failed to connect to scrcpy socket: %v", err)
	}
	defer conn.Close()
	log.Println("Connected to scrcpy server")

	width, height := readMetaData(conn)
	fmt.Printf("Width: %d\n", width)
	fmt.Printf("Height: %d\n", height)

	var data = &DecoderData{}
	if err := data.Allocate(width, height); err != nil {
		log.Println(fmt.Errorf("main: allocating decoder failed: %w", err))
		return
	}
	defer data.Free()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			err := handlePackets(conn, data)
			if err != nil {
				cancel()
				break
			}
		}
	}()
	showVideo(ctx, data)
}

func readMetaData(conn net.Conn) (width, height int) {
	var dummyByte = make([]byte, 1)
	_, err := conn.Read(dummyByte)
	if err != nil {
		fmt.Println(err)
		return width, height
	}
	fmt.Println(dummyByte)

	var deviceName = make([]byte, 64)
	_, err = conn.Read(deviceName)
	if err != nil {
		fmt.Println(err)
		return width, height
	}
	fmt.Println(deviceName)

	var codecID = make([]byte, 4)
	_, err = conn.Read(codecID)
	if err != nil {
		fmt.Println(err)
		return width, height
	}
	fmt.Println(codecID)

	var widthBuf = make([]byte, 4)
	_, err = conn.Read(widthBuf)
	if err != nil {
		fmt.Println(err)
		return width, height
	}
	fmt.Println(widthBuf)

	var heightBuf = make([]byte, 4)
	_, err = conn.Read(heightBuf)
	if err != nil {
		fmt.Println(err)
		return width, height
	}
	fmt.Println(heightBuf)

	width = int(binary.BigEndian.Uint32(widthBuf))
	height = int(binary.BigEndian.Uint32(heightBuf))
	return width, height
}

func handlePackets(
	conn net.Conn,
	data *DecoderData,
) error {
	if data == nil {
		return fmt.Errorf("decoder data is nil")
	}

	headerSize, err := io.ReadFull(conn, data.HeaderBuf)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if headerSize != 12 {
		fmt.Println("header size is not compatible")
		return nil
	}

	pts := binary.BigEndian.Uint64(data.HeaderBuf[0:8])
	fmt.Println(pts)
	packetSize := binary.BigEndian.Uint32(data.HeaderBuf[8:12])
	fmt.Println(packetSize)

	if pts <= 0 || int(packetSize) <= 0 {
		return nil
	}

	if BufSize < int(packetSize) {
		fmt.Printf("SIZE: %d not compatible\n", packetSize)
		return nil
	}

	size, err := io.ReadFull(conn, data.Buf[:packetSize])
	if err != nil {
		fmt.Printf("Connection closed or read error: %v\n", err)
		return err
	}

	if size < 0 || uint32(size) < packetSize {
		fmt.Printf("SIZE: %d not compatible\n", packetSize)
		return nil
	}

	var configSize = len(data.ConfigFrameBuf)
	var neededSpace = configSize + size
	if configSize > 0 {
		if neededSpace > len(data.Buf) {
			data.Buf = append(data.ConfigFrameBuf, data.Buf...)
		} else {
			copy(data.Buf[configSize:], data.Buf[:len(data.Buf)])
			copy(data.Buf[:configSize], data.ConfigFrameBuf)
		}
		data.ConfigFrameBuf = []byte{}
	}

	if pts&SCPacketFlagConfig != 0 {
		data.ConfigFrameBuf = append([]byte{}, data.Buf[:neededSpace]...)
		return nil
	}

	keyFrame := pts&SCPacketFlagKeyFrame != 0
	err = data.Decoder.Decode(
		data.Buf[:neededSpace],
		keyFrame,
		int64(pts&SCPacketPtsMask),
	)
	if err != nil {
		log.Println(fmt.Errorf("decoding packet failed: %w", err))
	}
	return nil
}

func frameToMat(data *DecoderData) (gocv.Mat, error) {
	img := data.Decoder.LatestYCbCr()
	if img == nil {
		return gocv.NewMat(), errors.New("no decoded frame available")
	}

	mat, err := gocv.ImageToMatRGB(img)
	if err != nil {
		mat.Close()
		return gocv.NewMat(), fmt.Errorf("failed to convert image to Mat: %w", err)
	}
	return mat, nil
}

func showVideo(
	ctx context.Context,
	data *DecoderData,
) {
	window := gocv.NewWindow("H264 Stream")
	defer window.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			img, err := frameToMat(data)
			if err != nil {
				img.Close()
				continue
			}

			if img.Empty() {
				log.Println("image is empty")
				img.Close()
				continue
			}

			rects, err := findAllRectangles(&img)
			if err != nil {
				fmt.Println(err)
				continue
			}
			drawRectangles(
				&img,
				rects...,
			)
			window.IMShow(img)
			img.Close()
			window.WaitKey(1)
		}
	}
}

func startScrcpy(serial string) error {
	scrcpyPath, err := downloadScrcpyServer()
	if err != nil {
		return err
	}

	var pushServerCmd = fmt.Sprintf(PushFile, serial, scrcpyPath)
	var forwardPortCmd = fmt.Sprintf(ForwardTCPPort, serial)
	var startScrcpyServerCmd = fmt.Sprintf(StartScrcpyServer, serial, ScrcpyVersion)

	_, err = executeCommand(pushServerCmd)
	if err != nil {
		return err
	}
	_, err = executeCommand(forwardPortCmd)
	if err != nil {
		return err
	}

	go executeCommand(startScrcpyServerCmd)
	return nil
}

func downloadScrcpyServer() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cacheDir, BasePath, ScrcpyDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	filePath := filepath.Join(dir, fmt.Sprintf(ScrcpyFileFormat, ScrcpyVersion))
	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("found scrcpy file: %s\n", filePath)
		return filePath, nil
	}

	downloadLink := fmt.Sprintf(ScrcpyLinkFormat, ScrcpyVersion, ScrcpyVersion)
	fmt.Printf("downloading scrcpy server: %s\n", downloadLink)
	resp, err := http.Get(downloadLink)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download %s: %s", downloadLink, resp.Status)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}
	fmt.Printf("scrcpy server downloaded: %s\n", filePath)
	return filePath, nil
}

func executeCommand(cmd string) (string, error) {
	if cmd == "" {
		return "", fmt.Errorf("cmd is empty")
	}
	cmdExec := exec.Command("bash", "-c", cmd)
	cmdExec.Stdin = os.Stdin
	cmdExec.Stderr = os.Stderr
	fmt.Println("-------")
	fmt.Printf("(%s): Start... ⏳\n", cmd)
	result, err := cmdExec.Output()
	if len(result) > 0 {
		var trimmedResult = strings.Trim(string(result), "\n")
		fmt.Printf("(%s): %s ✅\n", cmd, trimmedResult)
	} else {
		fmt.Printf("(%s): DONE ✅\n", cmd)
	}

	if err != nil {
		fmt.Printf("(%s): %s ❌\n", cmd, err.Error())
	}
	return string(result), err
}

// MinBorderDistance ...
const MinBorderDistance = 20

// Rectangle ...
type Rectangle struct {
	LeftX   int `json:"left_x"`
	RightX  int `json:"right_x"`
	TopY    int `json:"top_y"`
	BottomY int `json:"bottom_y"`
}

// ImgRectanglesToDomain ...
func ImgRectanglesToDomain(imgRectangles []image.Rectangle) []Rectangle {
	var rectangles []Rectangle
	for _, imgRectangle := range imgRectangles {
		rectangles = append(rectangles, Rectangle{
			LeftX:   imgRectangle.Min.X,
			RightX:  imgRectangle.Max.X,
			TopY:    imgRectangle.Min.Y,
			BottomY: imgRectangle.Max.Y,
		})
	}
	return rectangles
}

func findAllRectangles(
	img *gocv.Mat,
) ([]Rectangle, error) {
	imgRectangles, err := createRectangles(img)
	if err != nil {
		return []Rectangle{}, err
	}

	imgRectangles = filterRectangles(imgRectangles, img)
	rectangles := ImgRectanglesToDomain(imgRectangles)
	return rectangles, nil
}

func drawRectangles(
	img *gocv.Mat,
	rectangles ...Rectangle,
) {
	var redColor = color.RGBA{R: 255}
	for _, rect := range rectangles {
		imageRect := image.Rect(rect.LeftX, rect.TopY, rect.RightX, rect.BottomY)
		err := gocv.Rectangle(img, imageRect, redColor, 2)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func createRectangles(img *gocv.Mat) ([]image.Rectangle, error) {
	if img == nil {
		return []image.Rectangle{}, errors.New("img is nil")
	}
	grayImg := gocv.NewMat()
	defer grayImg.Close()
	err := gocv.CvtColor(*img, &grayImg, gocv.ColorBGRToGray)
	if err != nil {
		return []image.Rectangle{}, err
	}

	threshold := gocv.NewMat()
	defer threshold.Close()

	err = gocv.AdaptiveThreshold(
		grayImg,
		&threshold,
		255,
		gocv.AdaptiveThresholdGaussian,
		gocv.ThresholdBinary,
		11,
		2,
	)
	if err != nil {
		return []image.Rectangle{}, err
	}

	hierarchy := gocv.NewMat()
	defer hierarchy.Close()
	contours := gocv.FindContoursWithParams(
		threshold,
		&hierarchy,
		gocv.RetrievalList,
		gocv.ChainApproxSimple,
	)

	var rectangles []image.Rectangle
	for i := range contours.Size() {
		pts := contours.At(i)
		rect := gocv.BoundingRect(pts)
		rectangles = append(rectangles, rect)
	}
	return rectangles, nil
}

func filterRectangles(rects []image.Rectangle, img *gocv.Mat) []image.Rectangle {
	sort.Slice(rects, func(i, j int) bool {
		areaI := rects[i].Dx() * rects[i].Dy()
		areaJ := rects[j].Dx() * rects[j].Dy()
		return areaI > areaJ
	})

	filtered := make([]image.Rectangle, 0, len(rects))
	for _, rect := range rects {
		area := float64(rect.Dx() * rect.Dy())
		if area < 1000 || rect.Size().X >= img.Cols() || rect.Size().Y >= img.Rows() {
			continue
		}

		shouldKeep := true
		for _, larger := range filtered {
			currentArea := rect.Dx() * rect.Dy()
			largerArea := larger.Dx() * larger.Dy()

			if currentArea < largerArea && isCloseToBorder(rect, larger) {
				shouldKeep = false
				break
			}
		}

		if shouldKeep {
			filtered = append(filtered, rect)
		}
	}

	return filtered
}

func isCloseToBorder(inner, outer image.Rectangle) bool {
	leftDist := inner.Min.X - outer.Min.X
	rightDist := outer.Max.X - inner.Max.X
	topDist := inner.Min.Y - outer.Min.Y
	bottomDist := outer.Max.Y - inner.Max.Y

	return ((leftDist >= 0 && leftDist <= MinBorderDistance) ||
		(rightDist >= 0 && rightDist <= MinBorderDistance) ||
		(topDist >= 0 && topDist <= MinBorderDistance) ||
		(bottomDist >= 0 && bottomDist <= MinBorderDistance)) && inner.Overlaps(outer)
}
