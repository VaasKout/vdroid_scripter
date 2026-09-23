package cv

import (
	"android_vision_scripter/pkg/core/numutils"
	"android_vision_scripter/pkg/models"
	"android_vision_scripter/pkg/tesseract"
	"errors"
	"fmt"
	"image"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"gocv.io/x/gocv"
)

// Keyboard detection constants
const (
	KeyboardRowMatchRatio  = 0.5
	KeyboardRowMinMatches  = 2
	KeyboardRowSkip        = 1
	KeyboardGlyphMinRatio  = 0.6
	KeyboardRowHeightRatio = 0.5
	KeyboardRowGapRatio    = 0.6
	KeyboardKeyFillRatio   = 0.7
	KeyboardSpaceKeyWidth  = 2.0
	KeyboardZeroColumn     = 1
	KeyboardOcrWidth       = 540.0
	NumericWhitelist       = "0123456789"
)

// ErrKeyboardNotFound ...
var ErrKeyboardNotFound = errors.New("keyboard not found")

// KeyboardHandler ...
type KeyboardHandler interface {
	DetectKeyboard(img *gocv.Mat, locale string) (*models.Keyboard, error)
}

type glyph struct {
	char  rune
	upper bool
	rect  image.Rectangle
}

type matchedGlyph struct {
	index int
	glyph glyph
}

type keyboardRow struct {
	letters []rune
	glyphs  []matchedGlyph
	origin  float64
	y       int
}

func (c *cvImpl) DetectKeyboard(img *gocv.Mat, locale string) (*models.Keyboard, error) {
	layout, ok := KeyboardLayoutFor(locale)
	if !ok {
		return nil, fmt.Errorf("no keyboard layout for %s", locale)
	}
	if img == nil || img.Empty() {
		return nil, errors.New("keyboard img empty")
	}

	whitelist := ""
	if layout.Numeric {
		whitelist = NumericWhitelist
	}
	glyphs, err := c.readKeyboardGlyphs(img, TesseractLang(locale), whitelist)
	if err != nil {
		return nil, err
	}

	ocrRows := groupGlyphRows(glyphs)
	rows, err := matchLayoutRows(ocrRows, layout.Rows)
	if err != nil {
		return nil, err
	}

	pitch := keyPitch(rows)
	if pitch <= 0 {
		return nil, ErrKeyboardNotFound
	}
	setOrigins(rows, pitch, layout.Numeric)
	spacing := rowSpacing(rows)
	if spacing <= 0 {
		return nil, ErrKeyboardNotFound
	}

	if layout.Numeric {
		return buildNumericKeyboard(rows, pitch, spacing), nil
	}

	numbers := matchNumberRow(ocrRows, rows[0], pitch, spacing)
	keyboard := buildKeyboard(rows, numbers, pitch, spacing)
	keyboard.Shifted = shiftedByCase(rows, layout.CaseGlyphs)
	return keyboard, nil
}

func (c *cvImpl) readKeyboardGlyphs(img *gocv.Mat, lang string, whitelist string) ([]glyph, error) {
	scale := min(1, KeyboardOcrWidth/float64(img.Cols()))
	scaled := gocv.NewMat()
	defer scaled.Close()
	err := gocv.Resize(*img, &scaled, image.Point{}, scale, scale, gocv.InterpolationArea)
	if err != nil {
		return nil, err
	}

	edges, err := c.createEdges(&scaled)
	defer edges.Close()
	if err != nil {
		return nil, err
	}

	symbols, err := tesseract.RecognizeSymbols(
		edges.ToBytes(),
		edges.Cols(),
		edges.Rows(),
		edges.Cols(),
		lang,
		PsmBlock,
		OemText,
		whitelist,
	)
	if err != nil {
		return nil, err
	}

	glyphs := []glyph{}
	for _, symbol := range symbols {
		first, _ := utf8.DecodeRuneInString(strings.TrimSpace(symbol.Text))
		if first == utf8.RuneError {
			continue
		}
		glyphs = append(glyphs, glyph{
			char:  unicode.ToLower(first),
			upper: unicode.IsUpper(first),
			rect:  models.UnscaledRect(symbol.Rect, scale),
		})
	}
	return glyphs, nil
}

func groupGlyphRows(glyphs []glyph) [][]glyph {
	glyphs = dropSmallGlyphs(glyphs)
	if len(glyphs) == 0 {
		return nil
	}
	sort.Slice(glyphs, func(a, b int) bool {
		return models.CenterY(glyphs[a].rect) < models.CenterY(glyphs[b].rect)
	})

	gap := int(float64(medianHeight(glyphs)) * KeyboardRowGapRatio)
	rows := [][]glyph{}
	row := []glyph{glyphs[0]}
	for _, current := range glyphs[1:] {
		previous := row[len(row)-1]
		if models.CenterY(current.rect)-models.CenterY(previous.rect) > gap {
			rows = append(rows, sortedByX(row))
			row = []glyph{}
		}
		row = append(row, current)
	}
	rows = append(rows, sortedByX(row))
	return rows
}

func dropSmallGlyphs(glyphs []glyph) []glyph {
	if len(glyphs) == 0 {
		return glyphs
	}
	minHeight := int(float64(medianHeight(glyphs)) * KeyboardGlyphMinRatio)
	kept := []glyph{}
	for _, current := range glyphs {
		if current.rect.Dy() < minHeight {
			continue
		}
		kept = append(kept, current)
	}
	return kept
}

func sortedByX(row []glyph) []glyph {
	sort.Slice(row, func(a, b int) bool {
		return models.CenterX(row[a].rect) < models.CenterX(row[b].rect)
	})
	return row
}

func matchLayoutRows(ocrRows [][]glyph, layoutRows []string) ([]keyboardRow, error) {
	for bottom := len(ocrRows) - 1; bottom >= len(layoutRows)-1; bottom-- {
		rows, ok := matchRowsUpward(ocrRows, layoutRows, bottom)
		if ok {
			return rows, nil
		}
	}
	return nil, ErrKeyboardNotFound
}

func matchRowsUpward(ocrRows [][]glyph, layoutRows []string, bottom int) ([]keyboardRow, bool) {
	rows := make([]keyboardRow, len(layoutRows))
	from := bottom
	height := 0
	for layoutIndex := len(layoutRows) - 1; layoutIndex >= 0; layoutIndex-- {
		row, found, ok := matchRowNear(ocrRows, []rune(layoutRows[layoutIndex]), from, height)
		if !ok {
			return nil, false
		}
		rows[layoutIndex] = row
		from = found - 1
		height = matchedHeight(row)
	}
	return rows, true
}

func matchRowNear(ocrRows [][]glyph, letters []rune, from int, height int) (keyboardRow, int, bool) {
	for index := from; index >= 0 && index >= from-KeyboardRowSkip; index-- {
		matches := matchGlyphs(ocrRows[index], letters)
		if !enoughMatches(matches, letters) {
			continue
		}
		row := keyboardRow{letters: letters, glyphs: matches, y: medianMatchedY(matches)}
		if !similarHeight(matchedHeight(row), height) {
			continue
		}
		return row, index, true
	}
	return keyboardRow{}, 0, false
}

func similarHeight(height int, reference int) bool {
	if reference == 0 {
		return true
	}
	low := float64(min(height, reference))
	high := float64(max(height, reference))
	return low >= high*KeyboardRowHeightRatio
}

func matchGlyphs(row []glyph, letters []rune) []matchedGlyph {
	table := make([][]int, len(row)+1)
	for index := range table {
		table[index] = make([]int, len(letters)+1)
	}
	for a := len(row) - 1; a >= 0; a-- {
		for b := len(letters) - 1; b >= 0; b-- {
			if row[a].char == letters[b] {
				table[a][b] = table[a+1][b+1] + 1
				continue
			}
			table[a][b] = max(table[a+1][b], table[a][b+1])
		}
	}

	matches := []matchedGlyph{}
	a, b := 0, 0
	for a < len(row) && b < len(letters) {
		if row[a].char == letters[b] {
			matches = append(matches, matchedGlyph{index: b, glyph: row[a]})
			a++
			b++
			continue
		}
		if table[a+1][b] >= table[a][b+1] {
			a++
			continue
		}
		b++
	}
	return matches
}

func enoughMatches(matches []matchedGlyph, letters []rune) bool {
	if len(matches) < KeyboardRowMinMatches {
		return false
	}
	return float64(len(matches)) >= float64(len(letters))*KeyboardRowMatchRatio
}

func keyPitch(rows []keyboardRow) float64 {
	ratios := []float64{}
	for _, row := range rows {
		for pair := 1; pair < len(row.glyphs); pair++ {
			previous := row.glyphs[pair-1]
			current := row.glyphs[pair]
			dx := float64(models.CenterX(current.glyph.rect) - models.CenterX(previous.glyph.rect))
			ratios = append(ratios, dx/float64(current.index-previous.index))
		}
	}
	return numutils.MedianFloat(ratios)
}

func setOrigins(rows []keyboardRow, pitch float64, shared bool) {
	if shared {
		origin := numutils.MedianFloat(glyphOrigins(rows, pitch))
		for index := range rows {
			rows[index].origin = origin
		}
		return
	}
	for index := range rows {
		rows[index].origin = rowOrigin(rows[index], pitch)
	}
}

func rowOrigin(row keyboardRow, pitch float64) float64 {
	return numutils.MedianFloat(glyphOrigins([]keyboardRow{row}, pitch))
}

func glyphOrigins(rows []keyboardRow, pitch float64) []float64 {
	origins := []float64{}
	for _, row := range rows {
		for _, match := range row.glyphs {
			origins = append(origins, float64(models.CenterX(match.glyph.rect))-float64(match.index)*pitch)
		}
	}
	return origins
}

func rowSpacing(rows []keyboardRow) int {
	diffs := []int{}
	for index := 1; index < len(rows); index++ {
		diffs = append(diffs, rows[index].y-rows[index-1].y)
	}
	return numutils.MedianInt(diffs)
}

func matchNumberRow(
	ocrRows [][]glyph,
	top keyboardRow,
	pitch float64,
	spacing int,
) *keyboardRow {
	digits := []rune(NumberRow)
	minHeight := int(float64(matchedHeight(top)) * KeyboardGlyphMinRatio)
	for _, ocrRow := range ocrRows {
		matches := matchGlyphs(ocrRow, digits)
		if !enoughMatches(matches, digits) {
			continue
		}
		row := keyboardRow{letters: digits, glyphs: matches, y: medianMatchedY(matches)}
		if row.y >= top.y || top.y-row.y > spacing*3/2 {
			continue
		}
		if matchedHeight(row) < minHeight {
			continue
		}
		row.origin = rowOrigin(row, pitch)
		return &row
	}
	return nil
}

func buildKeyboard(
	rows []keyboardRow,
	numbers *keyboardRow,
	pitch float64,
	spacing int,
) *models.Keyboard {
	keyboard := models.NewKeyboard()
	for _, row := range rows {
		addRowKeys(keyboard, row, pitch, spacing)
	}
	if numbers != nil {
		addRowKeys(keyboard, *numbers, pitch, spacing)
	}

	left := rows[0].origin - pitch/2
	right := rows[0].origin + float64(len(rows[0].letters))*pitch - pitch/2
	for _, row := range rows[1:] {
		left = min(left, row.origin-pitch/2)
		right = max(right, row.origin+float64(len(row.letters))*pitch-pitch/2)
	}

	last := rows[len(rows)-1]
	firstLetterLeft := last.origin - pitch/2
	if firstLetterLeft-left >= pitch/2 {
		keyboard.Shift = keyRect((left+firstLetterLeft)/2, float64(last.y), firstLetterLeft-left, float64(spacing))
	}
	keyboard.Space = keyRect((left+right)/2, float64(last.y+spacing), pitch*KeyboardSpaceKeyWidth, float64(spacing))
	return keyboard
}

func buildNumericKeyboard(rows []keyboardRow, pitch float64, spacing int) *models.Keyboard {
	keyboard := models.NewKeyboard()
	for _, row := range rows {
		addRowKeys(keyboard, row, pitch, spacing)
	}

	last := rows[len(rows)-1]
	zeroX := last.origin + float64(KeyboardZeroColumn)*pitch
	keyboard.Keys['0'] = keyRect(zeroX, float64(last.y+spacing), pitch, float64(spacing))
	return keyboard
}

func addRowKeys(keyboard *models.Keyboard, row keyboardRow, pitch float64, spacing int) {
	for index, ch := range row.letters {
		keyboard.Keys[ch] = keyRect(row.origin+float64(index)*pitch, float64(row.y), pitch, float64(spacing))
	}
}

func keyRect(centerX float64, centerY float64, width float64, height float64) image.Rectangle {
	halfWidth := width * KeyboardKeyFillRatio / 2
	halfHeight := height * KeyboardKeyFillRatio / 2
	return image.Rect(
		int(centerX-halfWidth),
		int(centerY-halfHeight),
		int(centerX+halfWidth),
		int(centerY+halfHeight),
	)
}

func shiftedByCase(rows []keyboardRow, caseGlyphs string) bool {
	upper := 0
	lower := 0
	for _, row := range rows {
		for _, match := range row.glyphs {
			if !strings.ContainsRune(caseGlyphs, match.glyph.char) {
				continue
			}
			if match.glyph.upper {
				upper++
				continue
			}
			lower++
		}
	}
	return upper > lower
}

func medianMatchedY(matches []matchedGlyph) int {
	values := make([]int, 0, len(matches))
	for _, match := range matches {
		values = append(values, models.CenterY(match.glyph.rect))
	}
	return numutils.MedianInt(values)
}

func matchedHeight(row keyboardRow) int {
	values := make([]int, 0, len(row.glyphs))
	for _, match := range row.glyphs {
		values = append(values, match.glyph.rect.Dy())
	}
	return numutils.MedianInt(values)
}

func medianHeight(glyphs []glyph) int {
	values := make([]int, 0, len(glyphs))
	for _, current := range glyphs {
		values = append(values, current.rect.Dy())
	}
	return numutils.MedianInt(values)
}
