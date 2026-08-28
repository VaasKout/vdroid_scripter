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

	phraseMaxGapHeights = 3

	PsmText  = 11
	PsmChars = 7

	OemText  = 3
	OemChars = 0
)

// TesseractLocaleMap converts adb system locales
var TesseractLocaleMap = map[string]string{
	Numbers: "eng",
	Phone:   "eng",

	"af": "afr", "af-ZA": "afr", "afr": "afr",
	"am": "amh", "am-ET": "amh", "amh": "amh",
	"ar": "ara", "ar-AE": "ara", "ar-EG": "ara", "ar-SA": "ara", "ara": "ara",
	"as": "asm", "asm": "asm",
	"az": "aze", "az-AZ": "aze", "aze": "aze", "aze_cyrl": "aze_cyrl",
	"be": "bel", "be-BY": "bel", "bel": "bel",
	"bg": "bul", "bg-BG": "bul", "bul": "bul",
	"bn": "ben", "bn-BD": "ben", "bn-IN": "ben", "ben": "ben",
	"bo": "bod", "bod": "bod",
	"br": "bre", "bre": "bre",
	"bs": "bos", "bs-BA": "bos", "bos": "bos",
	"ca": "cat", "ca-ES": "cat", "cat": "cat",
	"ceb": "ceb",
	"chr": "chr",
	"co":  "cos", "cos": "cos",
	"cs": "ces", "cs-CZ": "ces", "ces": "ces",
	"cy": "cym", "cy-GB": "cym", "cym": "cym",
	"da": "dan", "da-DK": "dan", "dan": "dan",
	"de": "deu", "de-AT": "deu", "de-CH": "deu", "de-DE": "deu", "deu": "deu",
	"dv": "div", "div": "div",
	"dz": "dzo", "dzo": "dzo",
	"el": "ell", "el-GR": "ell", "ell": "ell",
	"en": "eng", "en-AU": "eng", "en-CA": "eng", "en-GB": "eng", "en-IN": "eng",
	"en-US": "eng", "eng": "eng", "enm": "enm",
	"eo": "epo", "epo": "epo",
	"equ": "equ",
	"es":  "spa", "es-AR": "spa", "es-ES": "spa", "es-MX": "spa", "es-US": "spa",
	"spa": "spa", "spa_old": "spa_old",
	"et": "est", "et-EE": "est", "est": "est",
	"eu": "eus", "eu-ES": "eus", "eus": "eus",
	"fa": "fas", "fa-IR": "fas", "fas": "fas",
	"fi": "fin", "fi-FI": "fin", "fin": "fin",
	"fil": "fil", "fil-PH": "fil",
	"fo": "fao", "fao": "fao",
	"fr": "fra", "fr-BE": "fra", "fr-CA": "fra", "fr-CH": "fra", "fr-FR": "fra",
	"fra": "fra", "frk": "frk", "frm": "frm",
	"fy": "fry", "fry": "fry",
	"ga": "gle", "ga-IE": "gle", "gle": "gle",
	"gd": "gla", "gla": "gla",
	"gl": "glg", "gl-ES": "glg", "glg": "glg",
	"grc": "grc",
	"gu":  "guj", "gu-IN": "guj", "guj": "guj",
	"ht": "hat", "hat": "hat",
	"he": "heb", "he-IL": "heb", "heb": "heb", "iw": "heb", "iw-IL": "heb",
	"hi": "hin", "hi-IN": "hin", "hin": "hin",
	"hr": "hrv", "hr-HR": "hrv", "hrv": "hrv",
	"hu": "hun", "hu-HU": "hun", "hun": "hun",
	"hy": "hye", "hy-AM": "hye", "hye": "hye",
	"id": "ind", "id-ID": "ind", "in": "ind", "in-ID": "ind", "ind": "ind",
	"is": "isl", "is-IS": "isl", "isl": "isl",
	"it": "ita", "it-CH": "ita", "it-IT": "ita", "ita": "ita", "ita_old": "ita_old",
	"iu": "iku", "iku": "iku",
	"ja": "jpn", "ja-JP": "jpn", "jpn": "jpn", "jpn_vert": "jpn_vert",
	"jv": "jav", "jav": "jav",
	"ka": "kat", "ka-GE": "kat", "kat": "kat", "kat_old": "kat_old",
	"kk": "kaz", "kk-KZ": "kaz", "kaz": "kaz",
	"km": "khm", "km-KH": "khm", "khm": "khm",
	"kmr": "kmr", "ku": "kmr",
	"kn": "kan", "kn-IN": "kan", "kan": "kan",
	"ko": "kor", "ko-KR": "kor", "kor": "kor", "kor_vert": "kor_vert",
	"ky": "kir", "ky-KG": "kir", "kir": "kir",
	"la": "lat", "lat": "lat",
	"lb": "ltz", "ltz": "ltz",
	"lo": "lao", "lo-LA": "lao", "lao": "lao",
	"lt": "lit", "lt-LT": "lit", "lit": "lit",
	"lv": "lav", "lv-LV": "lav", "lav": "lav",
	"mi": "mri", "mri": "mri",
	"mk": "mkd", "mk-MK": "mkd", "mkd": "mkd",
	"ml": "mal", "ml-IN": "mal", "mal": "mal",
	"mn": "mon", "mn-MN": "mon", "mon": "mon",
	"mr": "mar", "mr-IN": "mar", "mar": "mar",
	"ms": "msa", "ms-MY": "msa", "msa": "msa",
	"mt": "mlt", "mt-MT": "mlt", "mlt": "mlt",
	"my": "mya", "my-MM": "mya", "mya": "mya",
	"nb": "nor", "nb-NO": "nor", "nn": "nor", "nn-NO": "nor", "no": "nor",
	"nor": "nor",
	"ne":  "nep", "ne-NP": "nep", "nep": "nep",
	"nl": "nld", "nl-BE": "nld", "nl-NL": "nld", "nld": "nld",
	"oc": "oci", "oci": "oci",
	"or": "ori", "ori": "ori",
	"pa": "pan", "pa-IN": "pan", "pan": "pan",
	"pl": "pol", "pl-PL": "pol", "pol": "pol",
	"ps": "pus", "pus": "pus",
	"pt": "por", "pt-BR": "por", "pt-PT": "por", "por": "por",
	"qu": "que", "que": "que",
	"ro": "ron", "ro-RO": "ron", "ron": "ron",
	"ru": "rus", "ru-RU": "rus", "rus": "rus",
	"sa": "san", "san": "san",
	"sd": "snd", "snd": "snd",
	"si": "sin", "si-LK": "sin", "sin": "sin",
	"sk": "slk", "sk-SK": "slk", "slk": "slk",
	"sl": "slv", "sl-SI": "slv", "slv": "slv",
	"sq": "sqi", "sq-AL": "sqi", "sqi": "sqi",
	"sr": "srp", "sr-RS": "srp", "srp": "srp",
	"sr-Latn": "srp_latn", "srp_latn": "srp_latn",
	"su": "sun", "sun": "sun",
	"sv": "swe", "sv-SE": "swe", "swe": "swe",
	"sw": "swa", "sw-KE": "swa", "sw-TZ": "swa", "swa": "swa",
	"syr": "syr",
	"ta":  "tam", "ta-IN": "tam", "ta-LK": "tam", "tam": "tam",
	"tat": "tat", "tt": "tat",
	"te": "tel", "te-IN": "tel", "tel": "tel",
	"tg": "tgk", "tgk": "tgk",
	"th": "tha", "th-TH": "tha", "tha": "tha",
	"ti": "tir", "tir": "tir",
	"tl": "tgl", "tgl": "tgl",
	"to": "ton", "ton": "ton",
	"tr": "tur", "tr-TR": "tur", "tur": "tur",
	"ug": "uig", "uig": "uig",
	"uk": "ukr", "uk-UA": "ukr", "ukr": "ukr",
	"ur": "urd", "ur-IN": "urd", "ur-PK": "urd", "urd": "urd",
	"uz": "uzb", "uz-UZ": "uzb", "uzb": "uzb", "uzb_cyrl": "uzb_cyrl",
	"vi": "vie", "vi-VN": "vie", "vie": "vie",
	"yi": "yid", "ji": "yid", "yid": "yid",
	"yo": "yor", "yor": "yor",
	"zh": "chi_sim", "zh-CN": "chi_sim", "zh-Hans": "chi_sim", "zh-SG": "chi_sim",
	"chi_sim": "chi_sim", "chi_sim_vert": "chi_sim_vert",
	"zh-HK": "chi_tra", "zh-Hant": "chi_tra", "zh-MO": "chi_tra", "zh-TW": "chi_tra",
	"chi_tra": "chi_tra", "chi_tra_vert": "chi_tra_vert",
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
	for i := range contours.Size() {
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
	words := strings.Fields(text)
	if len(words) == 0 {
		return ocrArray
	}

	var ocrFilteredArray = []OCRResult{}
	for index := range ocrArray {
		phrase, ok := matchPhrase(ocrArray, index, words)
		if !ok {
			continue
		}
		ocrFilteredArray = append(ocrFilteredArray, phrase)
	}
	return ocrFilteredArray
}

func matchPhrase(ocrArray []OCRResult, start int, words []string) (OCRResult, bool) {
	if start+len(words) > len(ocrArray) {
		return OCRResult{}, false
	}

	merged := ocrArray[start]
	if !strings.EqualFold(merged.Text, words[0]) {
		return OCRResult{}, false
	}

	for i, word := range words {
		if i == 0 {
			continue
		}
		prev := ocrArray[start+i-1]
		next := ocrArray[start+i]
		if !strings.EqualFold(next.Text, word) {
			return OCRResult{}, false
		}
		if !isNextWordInLine(prev, next) {
			return OCRResult{}, false
		}

		merged.Rectangle.LeftX = min(merged.Rectangle.LeftX, next.Rectangle.LeftX)
		merged.Rectangle.TopY = min(merged.Rectangle.TopY, next.Rectangle.TopY)
		merged.Rectangle.RightX = max(merged.Rectangle.RightX, next.Rectangle.RightX)
		merged.Rectangle.BottomY = max(merged.Rectangle.BottomY, next.Rectangle.BottomY)
	}

	merged.Text = strings.Join(words, " ")
	return merged, true
}

func isNextWordInLine(prev OCRResult, next OCRResult) bool {
	prevRect := prev.Rectangle
	nextRect := next.Rectangle
	if nextRect.TopY >= prevRect.BottomY || nextRect.BottomY <= prevRect.TopY {
		return false
	}
	if nextRect.LeftX <= prevRect.LeftX {
		return false
	}

	height := prevRect.BottomY - prevRect.TopY
	gap := nextRect.LeftX - prevRect.RightX
	return gap <= height*phraseMaxGapHeights
}
