// Package tesseract ...
package tesseract

/*
#cgo pkg-config: tesseract
#include <stdlib.h>
#include <tesseract/capi.h>
*/
import "C"

import (
	"fmt"
	"image"
	"sync"
	"unsafe"
)

// Word ...
type Word struct {
	Text       string
	Confidence float32
	Rect       image.Rectangle
}

type engineKey struct {
	lang string
	oem  int
}

type engine struct {
	mu  sync.Mutex
	api *C.TessBaseAPI
}

var (
	enginesMu sync.Mutex
	engines   = map[engineKey]*engine{}
)

func getEngine(lang string, oem int) (*engine, error) {
	enginesMu.Lock()
	defer enginesMu.Unlock()

	key := engineKey{lang: lang, oem: oem}
	if result, ok := engines[key]; ok {
		return result, nil
	}

	api := C.TessBaseAPICreate()
	cLang := C.CString(lang)
	defer C.free(unsafe.Pointer(cLang))
	if C.TessBaseAPIInit2(api, nil, cLang, C.TessOcrEngineMode(oem)) != 0 {
		C.TessBaseAPIDelete(api)
		return nil, fmt.Errorf(
			"tesseract init failed for language %q (oem %d): is the traineddata installed?",
			lang, oem,
		)
	}

	created := &engine{api: api}
	engines[key] = created
	return created, nil
}

// Recognize ...
func Recognize(
	pixels []byte,
	width int,
	height int,
	stride int,
	lang string,
	psm int,
	oem int,
	whitelist string,
) ([]Word, error) {
	if width <= 0 || height <= 0 || stride < width {
		return nil, fmt.Errorf("invalid image %dx%d with stride %d", width, height, stride)
	}
	if len(pixels) < stride*height {
		return nil, fmt.Errorf("pixel buffer too small: %d < %d", len(pixels), stride*height)
	}

	eng, err := getEngine(lang, oem)
	if err != nil {
		return nil, err
	}

	eng.mu.Lock()
	defer eng.mu.Unlock()

	C.TessBaseAPISetPageSegMode(eng.api, C.TessPageSegMode(psm))

	cName := C.CString("tessedit_char_whitelist")
	defer C.free(unsafe.Pointer(cName))
	cWhitelist := C.CString(whitelist)
	defer C.free(unsafe.Pointer(cWhitelist))
	C.TessBaseAPISetVariable(eng.api, cName, cWhitelist)

	cPixels := C.CBytes(pixels[:stride*height])
	defer C.free(cPixels)
	C.TessBaseAPISetImage(
		eng.api,
		(*C.uchar)(cPixels),
		C.int(width), C.int(height), 1, C.int(stride),
	)
	if C.TessBaseAPIRecognize(eng.api, nil) != 0 {
		return nil, fmt.Errorf("tesseract recognize failed for language %q", lang)
	}

	iterator := C.TessBaseAPIGetIterator(eng.api)
	if iterator == nil {
		return []Word{}, nil
	}
	defer C.TessResultIteratorDelete(iterator)

	pageIterator := C.TessResultIteratorGetPageIterator(iterator)
	words := []Word{}
	for {
		cText := C.TessResultIteratorGetUTF8Text(iterator, C.RIL_WORD)
		if cText != nil {
			var left, top, right, bottom C.int
			C.TessPageIteratorBoundingBox(pageIterator, C.RIL_WORD, &left, &top, &right, &bottom)
			words = append(words, Word{
				Text:       C.GoString(cText),
				Confidence: float32(C.TessResultIteratorConfidence(iterator, C.RIL_WORD)),
				Rect:       image.Rect(int(left), int(top), int(right), int(bottom)),
			})
			C.TessDeleteText(cText)
		}
		if C.TessPageIteratorNext(pageIterator, C.RIL_WORD) == 0 {
			break
		}
	}
	return words, nil
}
