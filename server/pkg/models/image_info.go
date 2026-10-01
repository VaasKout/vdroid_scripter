package models

// ImageInfo ...
type ImageInfo struct {
	Density int `json:"density"`
}

// ScaleFor ...
func (i *ImageInfo) ScaleFor(density int) float64 {
	if i == nil || i.Density <= 0 || density <= 0 {
		return 1
	}
	return float64(density) / float64(i.Density)
}
