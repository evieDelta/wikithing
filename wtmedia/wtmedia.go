package wtmedia

import (
	"io"
	"net/url"
	"time"
)

type ObjectKind int

// Some recognised kinds for objects to have, some of these may allow additional query functions to be used
// (like resizing an image on the fly or requesting it as a jpeg instead of a png)
const (
	// Any generic data
	KindBinary ObjectKind = 0
	// Images and such, gives the ability to fetch it in another format or resize it on the fly
	KindImage ObjectKind = 1
)

// TypeMeta describes extra info about an object
type TypeMeta struct {
	Kind ObjectKind
	Mime string

	Meta map[string]string
}

// ObjectMeta is some various other object details
type ObjectMeta struct {
	Type TypeMeta

	Created time.Time
}

// QueryData specifies modifications or changes to be made to a query for kinds that support it
type QueryData struct {
	Extension string

	Values url.Values
}

// Writer is an interface for writing media objects
type Writer interface {
	Put(hash string, kind TypeMeta, data []byte) (err error)
	Rem(hash string) error
}

// Reader is an interface for getting media
type Reader interface {
	Get(hash string, query QueryData) (data io.ReadCloser, mime string, err error)
	GetMeta(hash string) (meta ObjectMeta, err error)
}

// MediaGet is an interface for defining a media store/get system
type MediaGet interface {
	Writer
	Reader
}
