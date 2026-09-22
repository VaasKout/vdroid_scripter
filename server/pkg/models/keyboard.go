package models

import (
	"fmt"
	"image"
	"unicode"
)

// Keyboard ...
type Keyboard struct {
	Shifted bool
	Shift   image.Rectangle
	Space   image.Rectangle
	Keys    map[rune]image.Rectangle
}

// NewKeyboard ...
func NewKeyboard() *Keyboard {
	return &Keyboard{Keys: map[rune]image.Rectangle{}}
}

// Key ...
func (k *Keyboard) Key(ch rune) (image.Rectangle, bool) {
	if k == nil {
		return image.Rectangle{}, false
	}
	rect, ok := k.Keys[ch]
	return rect, ok
}

// HasShift ...
func (k *Keyboard) HasShift() bool {
	return k != nil && !ImageRectIsEmpty(&k.Shift)
}

// ShiftNeeded ...
func (k *Keyboard) ShiftNeeded(ch rune) bool {
	return k != nil && unicode.IsLetter(ch) && unicode.IsUpper(ch) != k.Shifted
}

// KeyToPress ...
func (k *Keyboard) KeyToPress(ch rune) (image.Rectangle, error) {
	if k == nil {
		return image.Rectangle{}, fmt.Errorf("keyboard is nil")
	}
	if unicode.IsSpace(ch) {
		return k.Space, nil
	}
	key, found := k.Key(unicode.ToLower(ch))
	if !found {
		return image.Rectangle{}, fmt.Errorf("char %q not found on keyboard", ch)
	}
	return key, nil
}
