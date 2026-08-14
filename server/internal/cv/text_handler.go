package cv

import (
	"android_vision_scripter/pkg/core/file"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"gocv.io/x/gocv"
)

// Tesseract contants
const (
	DefaultOCRLanguage = "eng"
	Numbers            = "numbers"
	Phone              = "phone"
	MaxThreshHold      = 255

	darkRegionMinAreaRatio = 0.0025
	darkRegionMinAreaPx    = 2000.0
	darkRegionMinFillRatio = 0.5

	PsmText  = 11
	PsmChars = 7

	OemText  = 3
	OemChars = 0
)

// TesseractLocaleMap converts adb system locales
var TesseractLocaleMap = map[string]string{
	"rus":   "rus",
	"eng":   "eng",
	"ru":    "rus",
	"en":    "eng",
	"ru-RU": "rus",
	"en-US": "eng",
	Numbers: "eng",
	Phone:   "eng",
}

// OcrParams ...
type OcrParams struct {
	Text      string
	Lang      string
	Psm       int
	Oem       int
	WhiteList string
}

// TextHandler ...
type TextHandler interface {
	FindTextRectangles(
		img *gocv.Mat,
		dir string,
		params *OcrParams,
	) ([]OCRResult, error)
}

// InitOcrParams ...
func InitOcrParams(text string, lang string, psm int, oem int) *OcrParams {
	var ocrParams = new(OcrParams)
	ocrParams.Text = text
	ocrParams.Psm = psm
	ocrParams.Oem = oem

	if result, ok := TesseractLocaleMap[lang]; ok {
		ocrParams.Lang = result
	}

	if ocrParams.Lang == "" {
		ocrParams.Lang = DefaultOCRLanguage
	}

	if psm == 0 {
		ocrParams.Psm = 11
	}

	if lang == Numbers || lang == Phone {
		ocrParams.WhiteList = "-c tessedit_char_whitelist=0123456789"
	}

	return ocrParams
}

func (c *cvImpl) FindTextRectangles(
	img *gocv.Mat,
	dir string,
	params *OcrParams,
) ([]OCRResult, error) {
	if params == nil {
		return []OCRResult{}, errors.New("params are empty")
	}

	err := c.createEdges(img, dir)
	if err != nil {
		fmt.Println("Create edges error: " + err.Error())
		return []OCRResult{}, err
	}

	err = c.readTextFromEdgesImage(
		filepath.Join(dir, EdgesPng),
		params.Psm,
		params.Lang,
		params.Oem,
		params.WhiteList,
	)
	if err != nil {
		fmt.Println("ReadTextFromImage: " + err.Error())
		return []OCRResult{}, err
	}

	results := c.findRectangleInOcrJSON(filepath.Join(dir, OcrJSON), params.Text)
	return results, nil
}

func (c *cvImpl) createEdges(img *gocv.Mat, dir string) error {
	if img.Empty() {
		return errors.New("createEdges img empty")
	}
	gray := gocv.NewMat()
	defer gray.Close()
	err := gocv.CvtColor(*img, &gray, gocv.ColorBGRToGray)
	if err != nil {
		return err
	}

	edges := gocv.NewMat()
	defer edges.Close()
	gocv.Threshold(gray, &edges, 0, MaxThreshHold, gocv.ThresholdBinary|gocv.ThresholdOtsu)

	if gocv.CountNonZero(edges)*2 < edges.Rows()*edges.Cols() {
		gocv.BitwiseNot(edges, &edges)
	}

	err = invertDarkRegions(&edges)
	if err != nil {
		return err
	}

	var edgesPath = filepath.Join(dir, EdgesPng)
	ok := gocv.IMWrite(edgesPath, edges)
	if !ok {
		return errors.New("could not write " + EdgesPng)
	}
	return nil
}

func invertDarkRegions(edges *gocv.Mat) error {
	inverted := gocv.NewMat()
	defer inverted.Close()
	gocv.BitwiseNot(*edges, &inverted)

	contours := gocv.FindContours(inverted, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()

	regionMask := gocv.Zeros(edges.Rows(), edges.Cols(), gocv.MatTypeCV8UC1)
	defer regionMask.Close()

	minArea := max(float64(edges.Rows()*edges.Cols())*darkRegionMinAreaRatio, darkRegionMinAreaPx)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	found := false
	for i := 0; i < contours.Size(); i++ {
		rect := gocv.BoundingRect(contours.At(i))
		bboxArea := float64(rect.Dx() * rect.Dy())
		if bboxArea < minArea {
			continue
		}

		region := inverted.Region(rect)
		fillRatio := float64(gocv.CountNonZero(region)) / bboxArea
		region.Close()
		if fillRatio < darkRegionMinFillRatio {
			continue
		}

		err := gocv.DrawContours(&regionMask, contours, i, white, -1)
		if err != nil {
			return err
		}
		found = true
	}

	if !found {
		return nil
	}

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	err := gocv.Dilate(inverted, &inverted, kernel)
	if err != nil {
		return err
	}
	return inverted.CopyToWithMask(edges, regionMask)
}

func (c *cvImpl) readTextFromEdgesImage(
	imgPath string,
	psm int,
	lang string,
	oem int,
	whiteList string,
) error {
	dir, err := file.FindDirectoryOfFile(imgPath)
	if err != nil {
		fmt.Println(err)
		return err
	}

	cmd := fmt.Sprintf(
		"tesseract %s %s -l %s --psm %d --oem %d %s tsv",
		imgPath,
		filepath.Join(dir, OutputTsv),
		lang,
		psm,
		oem,
		whiteList,
	)
	_, err = c.cmdRunner.ExecuteCommand(cmd)
	if err != nil {
		fmt.Println("Error running Tesseract:", err)
		return err
	}

	// Read Tesseract output
	data, err := os.ReadFile(filepath.Join(dir, OutputTsv+".tsv"))
	if err != nil {
		fmt.Println("Error reading OCR output:", err)
		return err
	}
	results := TsvToOCRResult(string(data))

	// Save results as JSON
	ocrResultsFile, _ := os.Create(filepath.Join(dir, OcrJSON))
	defer ocrResultsFile.Close()
	err = json.NewEncoder(ocrResultsFile).Encode(results)
	return err
}

func (c *cvImpl) findRectangleInOcrJSON(osrJSONPath string, text string) []OCRResult {
	osrJSON, err := os.ReadFile(osrJSONPath)
	if err != nil {
		c.logAPI.Error(err.Error())
		return []OCRResult{}
	}

	var ocrArray = OCRJsonToArray(osrJSON)
	if text == "" {
		return ocrArray
	}

	var ocrFilteredArray = []OCRResult{}
	for _, ocr := range ocrArray {
		if strings.EqualFold(ocr.Text, text) {
			ocrFilteredArray = append(ocrFilteredArray, ocr)
		}
	}
	return ocrFilteredArray
}
