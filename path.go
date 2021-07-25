package wikithing

import (
	"strings"
	"unicode"
)

type Path struct{ p []string }

func (p Path) Path() string {
	return strings.Join(p.p, "/")
}

func (p Path) String() string {
	return strings.Join(p.p, ".")
}

const AllowedSpecialChars = "-()"

func ParsePath(s string) Path {
	p := Path{make([]string, 0)}

	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		switch r {
		case '.', '/', '\\', ' ':
			return '.'
		}
		if unicode.IsLetter(r) {
			return r
		}
		if unicode.IsNumber(r) {
			return r
		}
		if strings.ContainsAny(string(r), AllowedSpecialChars) {
			return r
		}

		return -1
	}, s)

	buf := strings.Builder{}
	for _, x := range s {
		if x != '.' {
			buf.WriteRune(x)
			continue
		}
		if buf.Len() == 0 {
			continue
		}
		p.p = append(p.p, buf.String())
		buf.Reset()
	}
	p.p = append(p.p, buf.String())

	return p
}
