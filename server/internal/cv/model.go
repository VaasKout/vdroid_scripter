package cv

import (
	"android_vision_scripter/pkg/models"
)

// OCRResult ...
type OCRResult struct {
	Text      string           `json:"text"`
	Rectangle models.Rectangle `json:"rectangle"`
}

// IsEmpty ...
func (o *OCRResult) IsEmpty() bool {
	return o == nil || o.Rectangle.IsEmpty() || o.Text == ""
}

// TemplateResult ...
type TemplateResult struct {
	Rectangle models.Rectangle
	Path      string
}
