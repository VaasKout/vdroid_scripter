package cv

import (
	"android_vision_scripter/pkg/models"
	"android_vision_scripter/pkg/tesseract"
	"errors"
	"image"
	"image/color"
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
	darkRegionMinContrast  = 30.0
	darkRegionMinTextRatio = 0.005

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
		ocrParams.WhiteList = "0123456789"
	}

	return ocrParams
}

func (c *cvImpl) FindTextRectangles(
	img *gocv.Mat,
	params *OcrParams,
) ([]OCRResult, error) {
	if params == nil {
		return []OCRResult{}, errors.New("params are empty")
	}

	edges, err := c.createEdges(img)
	defer edges.Close()
	if err != nil {
		return []OCRResult{}, err
	}

	words, err := tesseract.Recognize(
		edges.ToBytes(),
		edges.Cols(),
		edges.Rows(),
		edges.Cols(),
		params.Lang,
		params.Psm,
		params.Oem,
		params.WhiteList,
	)
	if err != nil {
		return []OCRResult{}, err
	}

	return filterResults(wordsToOCRResults(words), params.Text), nil
}

func (c *cvImpl) createEdges(img *gocv.Mat) (gocv.Mat, error) {
	if img.Empty() {
		return gocv.NewMat(), errors.New("createEdges img empty")
	}
	gray := gocv.NewMat()
	defer gray.Close()
	err := gocv.CvtColor(*img, &gray, gocv.ColorBGRToGray)
	if err != nil {
		return gocv.NewMat(), err
	}

	edges := gocv.NewMat()
	gocv.Threshold(gray, &edges, 0, MaxThreshHold, gocv.ThresholdBinary|gocv.ThresholdOtsu)

	err = rebinarizeDarkRegions(&edges, &gray)
	if err != nil {
		edges.Close()
		return gocv.NewMat(), err
	}

	if gocv.CountNonZero(edges)*2 < edges.Rows()*edges.Cols() {
		gocv.BitwiseNot(edges, &edges)
	}
	return edges, nil
}

func wordsToOCRResults(words []tesseract.Word) []OCRResult {
	results := []OCRResult{}
	for _, word := range words {
		if word.Text == "" {
			continue
		}
		results = append(results, OCRResult{
			Text: word.Text,
			Rectangle: models.Rectangle{
				LeftX:   word.Rect.Min.X,
				TopY:    word.Rect.Min.Y,
				RightX:  word.Rect.Max.X,
				BottomY: word.Rect.Max.Y,
			},
		})
	}
	return results
}

func rebinarizeDarkRegions(edges *gocv.Mat, gray *gocv.Mat) error {
	inverted := gocv.NewMat()
	defer inverted.Close()
	gocv.BitwiseNot(*edges, &inverted)

	contours := gocv.FindContours(inverted, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()

	labels := gocv.NewMat()
	defer labels.Close()
	gocv.ConnectedComponents(inverted, &labels)

	invertMask := gocv.Zeros(edges.Rows(), edges.Cols(), gocv.MatTypeCV8UC1)
	defer invertMask.Close()

	minArea := max(float64(edges.Rows()*edges.Cols())*darkRegionMinAreaRatio, darkRegionMinAreaPx)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	invertFound := false
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

		rebinarized, err := rebinarizeRegion(edges, gray, &labels, contours, i)
		if err != nil {
			return err
		}
		if rebinarized {
			continue
		}

		err = gocv.DrawContours(&invertMask, contours, i, white, -1)
		if err != nil {
			return err
		}
		invertFound = true
	}

	if !invertFound {
		return nil
	}

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	err := gocv.Dilate(inverted, &inverted, kernel)
	if err != nil {
		return err
	}
	return inverted.CopyToWithMask(edges, invertMask)
}

func rebinarizeRegion(
	edges *gocv.Mat,
	gray *gocv.Mat,
	labels *gocv.Mat,
	contours gocv.PointsVector,
	index int,
) (bool, error) {
	seed := contours.At(index).At(0)
	label := gocv.NewScalar(float64(labels.GetIntAt(seed.Y, seed.X)), 0, 0, 0)

	regionMask := gocv.NewMat()
	defer regionMask.Close()
	gocv.InRangeWithScalar(*labels, label, label, &regionMask)

	hist := gocv.NewMat()
	defer hist.Close()
	gocv.CalcHist([]gocv.Mat{*gray}, []int{0}, regionMask, &hist, []int{256}, []float64{0, 256}, false)

	threshold, contrast, textRatio := histOtsuStats(&hist)
	if contrast < darkRegionMinContrast {
		return false, nil
	}
	if textRatio < darkRegionMinTextRatio || textRatio >= 0.5 {
		return false, nil
	}

	binary := gocv.NewMat()
	defer binary.Close()
	gocv.Threshold(*gray, &binary, float32(threshold), MaxThreshHold, gocv.ThresholdBinary)
	return true, binary.CopyToWithMask(edges, regionMask)
}

func histOtsuStats(hist *gocv.Mat) (int, float64, float64) {
	bins := make([]float64, hist.Rows())
	total := 0.0
	weightedSum := 0.0
	for i := range bins {
		value := float64(hist.GetFloatAt(i, 0))
		bins[i] = value
		total += value
		weightedSum += float64(i) * value
	}
	if total == 0 {
		return 0, 0, 0
	}

	bestThreshold := 0
	bestVariance := -1.0
	weightBelow := 0.0
	sumBelow := 0.0
	for i, value := range bins {
		weightBelow += value
		if weightBelow == 0 {
			continue
		}
		weightAbove := total - weightBelow
		if weightAbove == 0 {
			break
		}
		sumBelow += float64(i) * value
		meanGap := (weightedSum-sumBelow)/weightAbove - sumBelow/weightBelow
		variance := weightBelow * weightAbove * meanGap * meanGap
		if variance > bestVariance {
			bestVariance = variance
			bestThreshold = i
		}
	}

	weightBelow = 0
	sumBelow = 0
	for i, value := range bins[:bestThreshold+1] {
		weightBelow += value
		sumBelow += float64(i) * value
	}
	weightAbove := total - weightBelow
	if weightBelow == 0 || weightAbove == 0 {
		return bestThreshold, 0, 0
	}
	contrast := (weightedSum-sumBelow)/weightAbove - sumBelow/weightBelow
	return bestThreshold, contrast, weightBelow / total
}

func filterResults(ocrArray []OCRResult, text string) []OCRResult {
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
