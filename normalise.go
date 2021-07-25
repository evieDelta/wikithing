package wikithing

import (
	"strings"
	"unicode"
)

// NormaliseText cleans up any excess control characters
func NormaliseText(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) {
			return r
		}
		return -1
	}, s)
}
