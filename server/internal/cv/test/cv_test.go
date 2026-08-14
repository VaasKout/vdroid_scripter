package test

import (
	"android_vision_scripter/config"
	"android_vision_scripter/internal/bashcmd"
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/logger"
	"android_vision_scripter/pkg/models"
	"fmt"
	"image"
	"path/filepath"
	"time"

	"testing"

	"gocv.io/x/gocv"
)

const (
	TestSerial   = "emulator-5554" //serial number of the device
	TestImage    = "./test.png"    //example template to compare zone on a screenshot
	TestLocale    = "eng"
	TestTextFile  = "./text_template.png"
	TestTextFile2 = "./text_template_2.png"
	TestSignWord  = "Sign"
)

func TestGetTextFromImage(t *testing.T) {
	var fileProps = &config.FilesProps{
		Logs: "./logs",
	}
	var logAPI = logger.New(logger.INFO, true)
	var filesDB = filesdb.New(fileProps)
	var cmdRunner = bashcmd.New(filesDB, logAPI)
	var cvAPI = cv.New(cmdRunner, logAPI)

	dir := filesDB.CreateLogsDir(TestSerial, filesdb.TesseractDir)
	testImage := TestTextFile

	img := gocv.IMRead(testImage, gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read screenshot image")
	}
	defer img.Close()

	ocrParams := cv.InitOcrParams("", TestLocale, cv.PsmText, cv.OemText)
	ocrResult, err := cvAPI.FindTextRectangles(&img, dir, ocrParams)
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
	var cmdRunner = bashcmd.New(filesDB, logAPI)
	var cvAPI = cv.New(cmdRunner, logAPI)

	dir := filesDB.CreateLogsDir(TestSerial, filesdb.TesseractDir)

	img := gocv.IMRead(TestTextFile2, gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read template image")
	}
	defer img.Close()

	ocrParams := cv.InitOcrParams(TestSignWord, TestLocale, cv.PsmText, cv.OemText)
	ocrResult, err := cvAPI.FindTextRectangles(&img, dir, ocrParams)
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
		t.Fatalf("expected 2 %q words, found %d", TestSignWord, len(rectangles))
	}
}

func TestGetTextFromScreenshot(t *testing.T) {
	var fileProps = &config.FilesProps{
		Logs: "./logs",
	}
	var logAPI = logger.New(logger.INFO, true)
	var filesDB = filesdb.New(fileProps)
	var cmdRunner = bashcmd.New(filesDB, logAPI)
	var cvAPI = cv.New(cmdRunner, logAPI)

	screenshot := cmdRunner.ScreenShot(TestSerial)
	if screenshot == "" {
		t.Fatal("screenshot is empty")
	}

	dir, err := file.FindDirectoryOfFile(screenshot)
	if err != nil {
		t.Fatal(err)
	}

	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read screenshot image")
	}
	defer img.Close()

	ocrParams := cv.InitOcrParams("", TestLocale, cv.PsmText, cv.OemText)
	ocrResult, err := cvAPI.FindTextRectangles(&img, dir, ocrParams)
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

func TestResetKeyboardKeys(t *testing.T) {
	var fileProps = &config.FilesProps{
		Logs: "./logs",
	}
	var logAPI = logger.New(logger.INFO, true)
	var filesDB = filesdb.New(fileProps)
	var cmdRunner = bashcmd.New(filesDB, logAPI)
	var cvAPI = cv.New(cmdRunner, logAPI)

	screenshot := cmdRunner.ScreenShot(TestSerial)
	if screenshot == "" {
		t.Fatal("screenshot is empty")
	}

	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read screenshot image")
	}
	defer img.Close()

	keyboardDir := filesDB.CreateKeyboardDir(TestSerial, TestLocale)
	tesseractDir := filesDB.CreateLogsDir(TestSerial, filesdb.TesseractDir)

	start := time.Now()
	ocrResult := cvAPI.ResetKeyboardKeys(keyboardDir, tesseractDir, screenshot, TestLocale, false)
	elapsed := time.Since(start)

	keyboardButtons := filesDB.GetFiles(keyboardDir)
	fmt.Printf("\nfound %d buttons for %d ms\n\n", len(keyboardButtons), elapsed.Milliseconds())

	var screenshotWithRects = filepath.Join(filesDB.CreateLogsDir(TestSerial), "screenshot.png")
	var rectangles = []image.Rectangle{}
	for _, result := range ocrResult {
		rectangles = append(rectangles, *result.Rectangle.ToImageRectangle())
	}
	cvAPI.DrawRectangles(img, rectangles, false)
	params := []int{gocv.IMWriteJpegQuality, 90}
	if ok := gocv.IMWriteWithParams(screenshotWithRects, img, params); !ok {
		fmt.Println("could not write image " + screenshotWithRects)
	}
}

func TestGetKeyboardKeys(t *testing.T) {
	var fileProps = &config.FilesProps{
		Logs: "./logs",
	}
	var logAPI = logger.New(logger.INFO, true)
	var filesDB = filesdb.New(fileProps)
	var cmdRunner = bashcmd.New(filesDB, logAPI)
	var cvAPI = cv.New(cmdRunner, logAPI)

	screenshot := cmdRunner.ScreenShot(TestSerial)
	if screenshot == "" {
		t.Fatal("screenshot is empty")
	}

	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read screenshot image")
	}
	defer img.Close()

	keyboardDir := filesDB.CreateKeyboardDir(TestSerial, TestLocale)
	keyboardButtons := filesDB.GetFiles(keyboardDir)

	if len(keyboardButtons) == 0 {
		t.Fatal(
			"run `TestResetKeyboardKeys` at first and make sure that keyboard is opened on the device screen",
		)
	}

	ocrResult := cvAPI.GetKeyboardKeys(keyboardButtons, img)
	t.Log(ocrResult)

	var screenshotWithRects = filepath.Join(filesDB.CreateLogsDir(TestSerial), "screenshot.png")
	var rectangles = []image.Rectangle{}
	for _, result := range ocrResult {
		rectangles = append(rectangles, *result.Rectangle.ToImageRectangle())
	}
	cvAPI.DrawRectangles(img, rectangles, false)
	if ok := gocv.IMWrite(screenshotWithRects, img); !ok {
		fmt.Println("could not write image " + screenshotWithRects)
	}
}

func TestDrawAllRectangles(t *testing.T) {
	var fileProps = &config.FilesProps{
		Logs: "./logs",
	}
	var logAPI = logger.New(logger.INFO, true)
	var filesDB = filesdb.New(fileProps)
	var cmdRunner = bashcmd.New(filesDB, logAPI)
	var cvAPI = cv.New(cmdRunner, logAPI)

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
	var cvAPI = cv.New(cmdRunner, logAPI)

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

	rectangle, err := cvAPI.FindImage(&img, TestImage)
	if err != nil {
		t.Fatal(err)
	}
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
