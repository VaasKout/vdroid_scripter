// Package yolo ...
package yolo

import (
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/logger"
	"android_vision_scripter/pkg/models"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gocv.io/x/gocv"
)

// Model files and detection params
const (
	ModelFile      = "best.onnx"
	ObjNames       = "obj.names"
	InputSize      = 640
	ScoreThreshold = 0.5
	NMSThreshold   = 0.45
	PadColor       = 114
)

// Detection ...
type Detection struct {
	Label      string
	Confidence float32
	Rectangle  models.Rectangle
}

// Yolo ...
type Yolo interface {
	DetectLabels(img gocv.Mat) []Detection
}

type yoloImpl struct {
	filesDB filesdb.FilesDB
	logAPI  *logger.Logger
	net     gocv.Net
	labels  []string
	mu      sync.Mutex
}

// New instance of Yolo
func New(filesDB filesdb.FilesDB, logAPI *logger.Logger) Yolo {
	var y = &yoloImpl{
		filesDB: filesDB,
		logAPI:  logAPI,
	}
	y.loadModel()
	return y
}

func (y *yoloImpl) loadModel() {
	dir := y.filesDB.CreateOnnxDir()
	if dir == "" {
		y.logAPI.Error("yolo dir not found")
		return
	}

	modelPath := filepath.Join(dir, ModelFile)
	if !file.Exists(modelPath) {
		y.logAPI.Error(fmt.Sprintf("onnx model not found: %s", modelPath))
		return
	}

	net := gocv.ReadNetFromONNX(modelPath)
	if net.Empty() {
		y.logAPI.Error(fmt.Sprintf("could not read onnx model: %s", modelPath))
		return
	}
	net.SetPreferableBackend(gocv.NetBackendDefault)
	net.SetPreferableTarget(gocv.NetTargetCPU)

	y.net = net
	y.labels = readLabels(filepath.Join(dir, ObjNames))
	y.logAPI.Info(fmt.Sprintf("loaded onnx model: %s ✅", modelPath))
}

func (y *yoloImpl) DetectLabels(img gocv.Mat) []Detection {
	if img.Empty() {
		return []Detection{}
	}

	padded, scale, padX, padY := letterbox(img, InputSize)
	defer padded.Close()

	blob := gocv.BlobFromImage(
		padded,
		1.0/255.0,
		image.Pt(InputSize, InputSize),
		gocv.NewScalar(0, 0, 0, 0),
		true,
		false,
	)
	defer blob.Close()

	y.mu.Lock()
	y.net.SetInput(blob, "")
	output := y.net.Forward("")
	y.mu.Unlock()
	defer output.Close()

	return y.parseOutput(output, scale, padX, padY, img.Cols(), img.Rows())
}

func (y *yoloImpl) parseOutput(
	output gocv.Mat,
	scale float64,
	padX int,
	padY int,
	originalWidth int,
	originalHeight int,
) []Detection {
	dims := output.Size()
	if len(dims) != 3 {
		return []Detection{}
	}

	reshaped := output.Reshape(1, dims[1])
	defer reshaped.Close()

	matrix := reshaped
	if dims[1] > dims[2] {
		transposed := gocv.NewMat()
		defer transposed.Close()
		gocv.Transpose(reshaped, &transposed)
		matrix = transposed
	}

	var numClasses = matrix.Rows() - 4
	var numBoxes = matrix.Cols()
	if numClasses <= 0 {
		return []Detection{}
	}

	var boxes = []image.Rectangle{}
	var scores = []float32{}
	var classes = []int{}

	for box := range numBoxes {
		bestScore, bestClass := highestScore(matrix, box, numClasses)
		if bestScore < ScoreThreshold || bestClass < 0 {
			continue
		}

		var centerX = float64(matrix.GetFloatAt(0, box))
		var centerY = float64(matrix.GetFloatAt(1, box))
		var width = float64(matrix.GetFloatAt(2, box))
		var height = float64(matrix.GetFloatAt(3, box))

		var left = (centerX - width/2 - float64(padX)) / scale
		var top = (centerY - height/2 - float64(padY)) / scale
		var right = (centerX + width/2 - float64(padX)) / scale
		var bottom = (centerY + height/2 - float64(padY)) / scale

		var rect = image.Rect(
			clamp(int(left), 0, originalWidth),
			clamp(int(top), 0, originalHeight),
			clamp(int(right), 0, originalWidth),
			clamp(int(bottom), 0, originalHeight),
		)
		boxes = append(boxes, rect)
		scores = append(scores, bestScore)
		classes = append(classes, bestClass)
	}

	indices := gocv.NMSBoxes(boxes, scores, ScoreThreshold, NMSThreshold)
	var detections = make([]Detection, 0, len(indices))
	for _, index := range indices {
		var box = boxes[index]
		detections = append(detections, Detection{
			Label:      y.labelName(classes[index]),
			Confidence: scores[index],
			Rectangle:  *models.ImgRectangleToDomain(&box),
		})
	}
	return detections
}

func highestScore(matrix gocv.Mat, box int, numClasses int) (float32, int) {
	var bestScore = float32(0)
	var bestClass = -1
	for class := range numClasses {
		var score = matrix.GetFloatAt(4+class, box)
		if score > bestScore {
			bestScore = score
			bestClass = class
		}
	}
	return bestScore, bestClass
}

func (y *yoloImpl) labelName(class int) string {
	if class >= 0 && class < len(y.labels) {
		return y.labels[class]
	}
	return strconv.Itoa(class)
}

func letterbox(img gocv.Mat, size int) (gocv.Mat, float64, int, int) {
	var width = img.Cols()
	var height = img.Rows()
	var scale = math.Min(float64(size)/float64(width), float64(size)/float64(height))
	var newWidth = int(math.Round(float64(width) * scale))
	var newHeight = int(math.Round(float64(height) * scale))

	resized := gocv.NewMat()
	defer resized.Close()
	gocv.Resize(img, &resized, image.Pt(newWidth, newHeight), 0, 0, gocv.InterpolationLinear)

	var padX = (size - newWidth) / 2
	var padY = (size - newHeight) / 2

	padded := gocv.NewMat()
	gocv.CopyMakeBorder(
		resized,
		&padded,
		padY,
		size-newHeight-padY,
		padX,
		size-newWidth-padX,
		gocv.BorderConstant,
		color.RGBA{R: PadColor, G: PadColor, B: PadColor, A: 0},
	)
	return padded, scale, padX, padY
}

func clamp(value int, low int, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func readLabels(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{}
	}

	var lines = strings.Split(strings.TrimSpace(string(data)), "\n")
	var labels = make([]string, 0, len(lines))
	for _, line := range lines {
		var trimmed = strings.TrimSpace(line)
		if trimmed != "" {
			labels = append(labels, trimmed)
		}
	}
	return labels
}
