package test

import (
	"android_vision_scripter/internal/server"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"testing"

	"gocv.io/x/gocv"
)

// Scrcpy paths
const (
	StartSessionPath = LocalURL + server.StartSession
)

func TestStartSession(t *testing.T) {
	var serialPath = fmt.Sprintf("{%s}", server.SerialKey)
	var requestURL = strings.ReplaceAll(StartSessionPath, serialPath, TestSerial)
	var data = ""
	makeHTTPRequest(
		requestURL,
		http.MethodGet,
		[]byte{},
		&data,
	)
	t.Log(data)
}

func TestConnectToScrcpy(t *testing.T) {
	var requestURL = fmt.Sprintf("%s?serial=%s", StartSessionPath, TestSerial)
	var data = map[string]any{}
	makeHTTPRequest(
		requestURL,
		http.MethodGet,
		[]byte{},
		&data,
	)
	t.Log(data)

	var videoStreamURL = fmt.Sprintf("tcp://127.0.0.1:%s", data["video_port"].(string))
	t.Logf("start listening %s", videoStreamURL)

	capture, err := gocv.OpenVideoCapture(videoStreamURL)
	if err != nil {
		log.Fatalf("error opening stream: %v", err)
	}

	go func() {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%s", data["control_port"].(string)))
		if err != nil {
			t.Logf("can't start socket conn: %s", err.Error())
		}

		buf := make([]byte, 32)
		for {
			_, err := conn.Read(buf)
			if err != nil {
				t.Logf("control disconnected: %s", err.Error())
				break
			}
		}

		t.Log("closing control connection... 🛑")
	}()

	defer capture.Close()
	fmt.Println("video connnection started")
	window := gocv.NewWindow("H264 Stream")
	defer window.Close()

	img := gocv.NewMat()
	defer img.Close()

	for {
		if ok := capture.Read(&img); !ok {
			continue
		}
		if img.Empty() {
			continue
		}

		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}

func TestReadRawVideoData(t *testing.T) {
	var requestURL = fmt.Sprintf("%s?serial=%s", StartSessionPath, TestSerial)
	var data = map[string]any{}
	makeHTTPRequest(
		requestURL,
		http.MethodGet,
		[]byte{},
		&data,
	)
	t.Log(data)

	var streamURL = fmt.Sprintf("127.0.0.1:%s", data["video_port"])
	conn, err := net.Dial("tcp", streamURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("connected to tcp://%s\n", streamURL)

	buf := make([]byte, 4096)
	for range 3 {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		var slice = buf[:n]
		fmt.Println(slice)
		fmt.Println(n)
	}
}
