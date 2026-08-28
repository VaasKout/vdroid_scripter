// Package main ...
package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
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

// Scrcpy properties
const (
	ScrcpyVersion = "3.3.4"

	ScrcpyLinkFormat = "https://github.com/Genymobile/scrcpy/releases/download/v%s/scrcpy-server-v%s"
	ScrcpyFileFormat = "scrcpy-v%s"
	BasePath         = "android_vision_scripter"
	ScrcpyDir        = "scrcpy"
)

// Adb commands
const (
	PushFile          = "adb -s %s push %s /data/local/tmp/scrcpy-server.jar"
	ForwardTCPPort    = "adb -s %s forward tcp:1234 localabstract:scrcpy"
	StartScrcpyServer = "adb -s %s shell CLASSPATH=/data/local/tmp/scrcpy-server.jar app_process / com.genymobile.scrcpy.Server %s log_level=verbose tunnel_forward=true audio=false control=false cleanup=false send_frame_meta=false"
)

func main() {
	var serial string
	if len(os.Args) > 1 {
		serial = strings.TrimSpace(os.Args[1])
	}
	if serial == "" {
		panic(fmt.Errorf("missing device serial: usage: go run . SERIAL"))
	}

	err := startScrcpy(serial)
	if err != nil {
		panic(err)
	}
	time.Sleep(1 * time.Second)
	var lastFrame = gocv.NewMat()
	defer lastFrame.Close()

	go createVideoConnection(&lastFrame)
	showVideo(&lastFrame)
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

func showVideo(
	lastFrame *gocv.Mat,
) {
	window := gocv.NewWindow("H264 Stream")
	defer window.Close()

	for {
		window.IMShow(*lastFrame)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}

func createVideoConnection(
	lastFrame *gocv.Mat,
) {
	streamURL := "tcp://127.0.0.1:1234"

	capture, err := gocv.OpenVideoCaptureWithAPIParams(
		streamURL,
		gocv.VideoCaptureFFmpeg,
		[]gocv.VideoCaptureProperties{
			54, 0, //CAP_PROP_READ_TIMEOUT_MSEC
			// 38, 1024,
			52, 1, //CAP_PROP_HW_ACCELERATION_USE_OPENCL
			// 8, -1, //CAP_PROP_FORMAT
		},
	)

	if err != nil {
		log.Fatalf("error opening stream: %v", err)
	}
	defer capture.Close()

	fmt.Printf("CODEC: %s\n", capture.CodecString())
	fmt.Printf("CAP_PROP_HW_ACCELERATION_USE_OPENCL: %f\n", capture.Get(52))
	fmt.Printf("CAP_PROP_N_THREADS: %f\n", capture.Get(70))
	fmt.Printf("CAP_PROP_FORMAT: %f\n", capture.Get(8))
	fmt.Printf("CAP_PROP_BUFFERSIZE: %f\n", capture.Get(38))
	fmt.Printf("CAP_PROP_CODEC_PIXEL_FORMAT: %f\n", capture.Get(46))
	fmt.Printf("CAP_PROP_ORIENTATION_AUTO: %f\n", capture.Get(49))
	fmt.Printf("CAP_PROP_CODEC_EXTRADATA_INDEX: %f\n", capture.Get(68))

	fmt.Println("video connnection started")

	for {
		if ok := capture.Read(lastFrame); !ok {
			continue
		}

		// fmt.Printf("CAP_PROP_LRF_HAS_KEY_FRAME: %f\n", capture.Get(67))
		// fmt.Printf("CAP_PROP_CODEC_EXTRADATA_INDEX: %f\n", capture.Get(68))
		// fmt.Printf("CAP_PROP_FRAME_TYPE: %f\n", capture.Get(69))
		// fmt.Printf("CAP_PROP_PTS: %f\n", capture.Get(71))
		// fmt.Printf("CAP_PROP_DTS_DELAY: %f\n", capture.Get(72))

		rects, err := findAllRectangles(lastFrame)
		if err != nil {
			fmt.Println(err)
			continue
		}
		drawRectangles(
			lastFrame,
			rects...,
		)
	}
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
		gocv.AdaptiveThresholdGaussian, // Often better than Mean for icons
		gocv.ThresholdBinary,
		11, // Larger block size
		2,  // Higher C value
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
	// sort rectangles from big to small
	sort.Slice(rects, func(i, j int) bool {
		areaI := rects[i].Dx() * rects[i].Dy()
		areaJ := rects[j].Dx() * rects[j].Dy()
		return areaI > areaJ
	})

	filtered := make([]image.Rectangle, 0, len(rects))
	for _, rect := range rects {
		// remove too small rectangles and rectangles with borders on the edge of the screen
		area := float64(rect.Dx() * rect.Dy())
		if area < 1000 || rect.Size().X >= img.Cols() || rect.Size().Y >= img.Rows() {
			continue
		}

		shouldKeep := true
		for _, larger := range filtered {
			currentArea := rect.Dx() * rect.Dy()
			largerArea := larger.Dx() * larger.Dy()

			//check rectangle only if it is smaller than some is already in filtered
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
