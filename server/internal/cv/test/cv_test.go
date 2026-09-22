package test

import (
	"android_vision_scripter/config"
	"android_vision_scripter/internal/bashcmd"
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/logger"
	"android_vision_scripter/pkg/models"
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"time"

	"testing"

	"gocv.io/x/gocv"
)

const (
	TestSerial     = "emulator-5554" //serial number of the device
	TestImage      = "./test.png"    //example template to compare zone on a screenshot
	TestLocale     = "eng"
	TestTextFile   = "./text_template.png"
	TestTextFile2  = "./text_template_2.png"
	TestTextFile3  = "./text_template_3.png"
	TestSignPhrase = "Sign In"
)

func TestGetTextFromImage(t *testing.T) {
	var fileProps = &config.FilesProps{
		Logs: "./logs",
	}
	var logAPI = logger.New(logger.INFO, true)
	var filesDB = filesdb.New(fileProps)
	var cvAPI = cv.New(logAPI)
	dir := filesDB.CreateLogsDir(TestSerial)

	testImage := TestTextFile3

	img := gocv.IMRead(testImage, gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read screenshot image")
	}
	defer img.Close()

	ocrParams := cv.InitOcrParams("", TestLocale, cv.PsmText, cv.OemText)
	ocrResult, err := cvAPI.FindTextRectangles(&img, ocrParams)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ocrResult)
	var rectangles = []image.Rectangle{}
	for _, ocr := range ocrResult {
		var imgRect = ocr.Rectangle.ToImageRectangle()
		if imgRect == nil || models.ImageRectIsEmpty(imgRect) {
			continue
		}
		rectangles = append(rectangles, *imgRect)
	}

	err = cvAPI.DrawRectangles(img, rectangles, false)
	if err != nil {
		t.Fatal(err)
	}
	params := []int{gocv.IMWriteJpegQuality, 90}
	if ok := gocv.IMWriteWithParams(filepath.Join(dir, testImage), img, params); !ok {
		fmt.Println("could not write image " + testImage)
	}
}

func TestFindSignText(t *testing.T) {
	var fileProps = &config.FilesProps{
		Logs: "./logs",
	}
	var logAPI = logger.New(logger.INFO, true)
	var filesDB = filesdb.New(fileProps)
	var cvAPI = cv.New(logAPI)
	dir := filesDB.CreateLogsDir(TestSerial)

	img := gocv.IMRead(TestTextFile2, gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read template image")
	}
	defer img.Close()

	ocrParams := cv.InitOcrParams(TestSignPhrase, TestLocale, cv.PsmText, cv.OemText)
	ocrResult, err := cvAPI.FindTextRectangles(&img, ocrParams)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ocrResult)
	var rectangles = []image.Rectangle{}
	for _, ocr := range ocrResult {
		var imgRect = ocr.Rectangle.ToImageRectangle()
		if imgRect == nil || models.ImageRectIsEmpty(imgRect) {
			continue
		}
		rectangles = append(rectangles, *imgRect)
	}

	err = cvAPI.DrawRectangles(img, rectangles, false)
	if err != nil {
		t.Fatal(err)
	}
	params := []int{gocv.IMWriteJpegQuality, 90}
	if ok := gocv.IMWriteWithParams(filepath.Join(dir, TestTextFile2), img, params); !ok {
		fmt.Println("could not write image " + TestTextFile2)
	}

	if len(rectangles) != 2 {
		t.Fatalf("expected 2 %q phrases, found %d", TestSignPhrase, len(rectangles))
	}
}

func TestGetTextFromScreenshot(t *testing.T) {
	var fileProps = &config.FilesProps{
		Logs: "./logs",
	}
	var logAPI = logger.New(logger.INFO, true)
	var filesDB = filesdb.New(fileProps)
	var cmdRunner = bashcmd.New(filesDB, logAPI)
	var cvAPI = cv.New(logAPI)

	screenshot := cmdRunner.ScreenShot(TestSerial)
	if screenshot == "" {
		t.Fatal("screenshot is empty")
	}

	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read screenshot image")
	}
	defer img.Close()

	ocrParams := cv.InitOcrParams("", TestLocale, cv.PsmText, cv.OemText)
	ocrResult, err := cvAPI.FindTextRectangles(&img, ocrParams)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ocrResult)
	var rectangles = []image.Rectangle{}
	for _, ocr := range ocrResult {
		var imgRect = ocr.Rectangle.ToImageRectangle()
		if imgRect == nil || models.ImageRectIsEmpty(imgRect) {
			continue
		}
		rectangles = append(rectangles, *imgRect)
	}

	err = cvAPI.DrawRectangles(img, rectangles, false)
	if err != nil {
		t.Fatal(err)
	}
	params := []int{gocv.IMWriteJpegQuality, 90}
	if ok := gocv.IMWriteWithParams(screenshot, img, params); !ok {
		fmt.Println("could not write image " + screenshot)
	}
}

func TestDrawAllRectangles(t *testing.T) {
	var fileProps = &config.FilesProps{
		Logs: "./logs",
	}
	var logAPI = logger.New(logger.INFO, true)
	var filesDB = filesdb.New(fileProps)
	var cmdRunner = bashcmd.New(filesDB, logAPI)
	var cvAPI = cv.New(logAPI)

	screenshot := cmdRunner.ScreenShot(TestSerial)
	if screenshot == "" {
		t.Fatal("screenshot is empty")
	}
	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read screenshot image")
	}
	defer img.Close()

	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)

	rectangles, err := cvAPI.FindAllRectangles(&gray)
	if err != nil {
		t.Fatal(err)
	}

	if len(rectangles) == 0 {
		t.Fatal("no rectangles found")
	}

	t.Logf("rects len: %d", len(rectangles))
	t.Log(rectangles)

	err = cvAPI.DrawRectangles(img, rectangles, false)
	if err != nil {
		t.Fatal(err)
	}
	var screenshotWithRects = filepath.Join(filesDB.CreateLogsDir(TestSerial), "screenshot.png")
	params := []int{gocv.IMWriteJpegQuality, 90}
	if ok := gocv.IMWriteWithParams(screenshotWithRects, img, params); !ok {
		fmt.Println("could not write image " + screenshot)
	}
}

func TestFindTemplate(t *testing.T) {
	var fileProps = &config.FilesProps{
		Logs: "./logs",
	}
	var logAPI = logger.New(logger.INFO, true)
	var filesDB = filesdb.New(fileProps)
	var cmdRunner = bashcmd.New(filesDB, logAPI)
	var cvAPI = cv.New(logAPI)

	screenshot := cmdRunner.ScreenShot(TestSerial)
	if screenshot == "" {
		t.Fatal("screenshot is empty")
	}
	fmt.Println(screenshot)

	start := time.Now()
	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read screenshot image")
	}
	defer img.Close()

	rectangles, err := cvAPI.FindImages(&img, TestImage)
	if err != nil {
		t.Fatal(err)
	}
	if len(rectangles) == 0 {
		t.Fatal("template not found")
	}
	rectangle := &rectangles[0]
	t.Log(rectangle)

	elapsed := time.Since(start)
	fmt.Printf("\nfound template for %d ms\n\n", elapsed.Milliseconds())

	err = cvAPI.DrawRectangles(img, []image.Rectangle{*rectangle}, false)
	if err != nil {
		t.Fatal(err)
	}
	var screenshotWithRects = filepath.Join(filesDB.CreateLogsDir(TestSerial), "screenshot.png")
	params := []int{gocv.IMWriteJpegQuality, 90}
	if ok := gocv.IMWriteWithParams(screenshotWithRects, img, params); !ok {
		fmt.Println("could not write image " + screenshot)
	}
}

func TestDetectKeyboard(t *testing.T) {
	var fileProps = &config.FilesProps{
		Logs: "./logs",
	}
	var logAPI = logger.New(logger.INFO, true)
	var filesDB = filesdb.New(fileProps)
	var cmdRunner = bashcmd.New(filesDB, logAPI)
	var cvAPI = cv.New(logAPI)

	screenshot := cmdRunner.ScreenShot(TestSerial)
	if screenshot == "" {
		t.Fatal("screenshot is empty")
	}
	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read screenshot image")
	}
	defer img.Close()

	start := time.Now()
	keyboard, err := cvAPI.DetectKeyboard(&img, TestLocale)
	if err != nil {
		t.Fatal(err)
	}
	keys := keyboard.Keys
	t.Logf("detected %d keys in %d ms, shifted: %v", len(keys), time.Since(start).Milliseconds(), keyboard.Shifted)

	var blueColor = color.RGBA{B: 255, A: 255}
	rectangles := []image.Rectangle{keyboard.Space}
	if keyboard.HasShift() {
		rectangles = append(rectangles, keyboard.Shift)
	}
	for ch, rect := range keys {
		rectangles = append(rectangles, rect)
		gocv.PutText(&img, string(ch), image.Pt(rect.Min.X, rect.Min.Y), gocv.FontHersheySimplex, 1, blueColor, 2)
	}

	err = cvAPI.DrawRectangles(img, rectangles, false)
	if err != nil {
		t.Fatal(err)
	}
	var screenshotWithKeys = filepath.Join(filesDB.CreateLogsDir(TestSerial), "keyboard.png")
	if ok := gocv.IMWrite(screenshotWithKeys, img); !ok {
		t.Fatal("could not write image " + screenshotWithKeys)
	}
}
