package cv

import (
	"android_vision_scripter/pkg/core/file"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gocv.io/x/gocv"
)

// Tesseract contants
const (
	DefaultOCRLanguage = "eng"
	Numbers            = "nums"
	WhiteTheme         = 230
	BlackTheme         = 120
	MaxThreshHold      = 255

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
}

// OcrParams ...
type OcrParams struct {
	Text      string
	Lang      string
	Psm       int
	Theme     float32
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
func InitOcrParams(text string, lang string, psm int, oem int, theme float32) *OcrParams {
	var ocrParams = new(OcrParams)
	ocrParams.Text = text
	ocrParams.Theme = theme
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

	if lang == Numbers {
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

	err := c.createEdges(img, dir, params.Theme)
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

	result := c.findRectangleInOcrJSON(filepath.Join(dir, OcrJSON), params.Text)
	return result, nil
}

func (c *cvImpl) createEdges(img *gocv.Mat, dir string, thresh float32) error {
	if img.Empty() {
		return errors.New("createEdges img empty")
	}
	edges := gocv.NewMat()
	defer edges.Close()

	gray := gocv.NewMat()
	defer gray.Close()
	err := gocv.CvtColor(*img, &gray, gocv.ColorBGRToGray)
	if err != nil {
		return err
	}
	gocv.Threshold(gray, &edges, thresh, MaxThreshHold, gocv.ThresholdBinary)

	var edgesPath = filepath.Join(dir, EdgesPng)
	ok := gocv.IMWrite(edgesPath, edges)
	if !ok {
		return errors.New("could not write " + EdgesPng)
	}
	return nil
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
