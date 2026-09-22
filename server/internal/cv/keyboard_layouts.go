package cv

// KeyboardLayout ...
type KeyboardLayout struct {
	Rows       []string
	CaseGlyphs string
}

// Letters whose upper and lower case glyphs differ in shape, not only in size
const (
	LatinCaseGlyphs    = "abdefghmnqrt"
	CyrillicCaseGlyphs = "абеё"
	GreekCaseGlyphs    = "αβγδεηλμνρσω"
	NumberRow          = "1234567890"
)

var (
	qwertyRows = []string{"qwertyuiop", "asdfghjkl", "zxcvbnm"}
	qwertzRows = []string{"qwertzuiop", "asdfghjkl", "yxcvbnm"}
	balkanRows = []string{"qwertzuiopšđ", "asdfghjklčć", "yxcvbnmž"}
	nordicRows = []string{"qwertyuiopå", "asdfghjklöä", "zxcvbnm"}
)

var keyboardLayouts = map[string]KeyboardLayout{
	"eng":      {Rows: qwertyRows, CaseGlyphs: LatinCaseGlyphs},
	"ita":      {Rows: qwertyRows, CaseGlyphs: LatinCaseGlyphs},
	"nld":      {Rows: qwertyRows, CaseGlyphs: LatinCaseGlyphs},
	"pol":      {Rows: qwertyRows, CaseGlyphs: LatinCaseGlyphs},
	"por":      {Rows: qwertyRows, CaseGlyphs: LatinCaseGlyphs},
	"ron":      {Rows: qwertyRows, CaseGlyphs: LatinCaseGlyphs},
	"lit":      {Rows: qwertyRows, CaseGlyphs: LatinCaseGlyphs},
	"lav":      {Rows: qwertyRows, CaseGlyphs: LatinCaseGlyphs},
	"est":      {Rows: qwertyRows, CaseGlyphs: LatinCaseGlyphs},
	"cat":      {Rows: qwertyRows, CaseGlyphs: LatinCaseGlyphs},
	"spa":      {Rows: []string{"qwertyuiop", "asdfghjklñ", "zxcvbnm"}, CaseGlyphs: LatinCaseGlyphs},
	"fra":      {Rows: []string{"azertyuiop", "qsdfghjklm", "wxcvbn"}, CaseGlyphs: LatinCaseGlyphs},
	"deu":      {Rows: []string{"qwertzuiopü", "asdfghjklöä", "yxcvbnm"}, CaseGlyphs: LatinCaseGlyphs},
	"ces":      {Rows: qwertzRows, CaseGlyphs: LatinCaseGlyphs},
	"slk":      {Rows: qwertzRows, CaseGlyphs: LatinCaseGlyphs},
	"hun":      {Rows: []string{"qwertzuiopőú", "asdfghjkléáű", "yxcvbnm"}, CaseGlyphs: LatinCaseGlyphs},
	"hrv":      {Rows: balkanRows, CaseGlyphs: LatinCaseGlyphs},
	"slv":      {Rows: balkanRows, CaseGlyphs: LatinCaseGlyphs},
	"srp_latn": {Rows: balkanRows, CaseGlyphs: LatinCaseGlyphs},
	"bos":      {Rows: balkanRows, CaseGlyphs: LatinCaseGlyphs},
	"swe":      {Rows: nordicRows, CaseGlyphs: LatinCaseGlyphs},
	"fin":      {Rows: nordicRows, CaseGlyphs: LatinCaseGlyphs},
	"nor":      {Rows: []string{"qwertyuiopå", "asdfghjkløæ", "zxcvbnm"}, CaseGlyphs: LatinCaseGlyphs},
	"dan":      {Rows: []string{"qwertyuiopå", "asdfghjklæø", "zxcvbnm"}, CaseGlyphs: LatinCaseGlyphs},
	"isl":      {Rows: []string{"qwertyuiopð", "asdfghjklæ", "zxcvbnmþ"}, CaseGlyphs: LatinCaseGlyphs},
	"tur":      {Rows: []string{"qwertyuıopğü", "asdfghjklşi", "zxcvbnmöç"}, CaseGlyphs: LatinCaseGlyphs},
	"rus":      {Rows: []string{"йцукенгшщзх", "фывапролджэ", "ячсмитьбю"}, CaseGlyphs: CyrillicCaseGlyphs},
	"ukr":      {Rows: []string{"йцукенгшщзхї", "фівапролджє", "ячсмитьбю"}, CaseGlyphs: CyrillicCaseGlyphs},
	"bel":      {Rows: []string{"йцукенгшўзх", "фывапролджэ", "ячсмітьбю"}, CaseGlyphs: CyrillicCaseGlyphs},
	"bul":      {Rows: []string{"уеишщксдзц", "ьяаожгтнвмч", "юйъэфхпрлб"}, CaseGlyphs: CyrillicCaseGlyphs},
	"srp":      {Rows: []string{"љњертзуиопшђ", "асдфгхјклчћж", "ѕџцвбнм"}, CaseGlyphs: CyrillicCaseGlyphs},
	"mkd":      {Rows: []string{"љњертѕуиопшѓ", "асдфгхјклчќж", "зџцвбнм"}, CaseGlyphs: CyrillicCaseGlyphs},
	"ell":      {Rows: []string{"ςερτυθιοπ", "ασδφγηξκλ", "ζχψωβνμ"}, CaseGlyphs: GreekCaseGlyphs},
}

// KeyboardLayoutFor ...
func KeyboardLayoutFor(lang string) (KeyboardLayout, bool) {
	layout, ok := keyboardLayouts[lang]
	return layout, ok
}
