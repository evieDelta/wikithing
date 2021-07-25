package wterr

import "fmt"

type ErrType uint

func (e ErrType) String() (s string) {
	switch e {
	default:
		fallthrough
	case ErrUnknown:
		s = "generic error"
	case ErrError:
		s = "error"
	case ErrAuthFailed:
		s = "unauthenticated"
	case ErrInvalidInput:
		s = "invalid input"
	case ErrUnsupported:
		s = "unsupported"
	}
	return s
}

const (
	ErrUnknown ErrType = iota
	ErrError
	ErrAuthFailed
	ErrInvalidInput
	ErrUnsupported
)

type Err struct {
	Type ErrType
	Str  string
}

func (e Err) Error() string { return e.Type.String() + ": " + e.Str }

func NewRaw(kind ErrType, i string) error {
	return Err{
		kind,
		i,
	}
}

func New(kind ErrType, i ...interface{}) error {
	s := fmt.Sprint(i...)
	return Err{
		kind,
		s,
	}
}

func Newln(kind ErrType, i ...interface{}) error {
	s := fmt.Sprintln(i...)
	return Err{
		kind,
		s[:len(s)-1],
	}
}

func Newf(kind ErrType, f string, i ...interface{}) error {
	s := fmt.Sprintf(f, i...)
	return Err{
		kind,
		s,
	}
}
