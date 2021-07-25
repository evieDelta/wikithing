package wtfs

import (
	"errors"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
)

const Extension = ".json"

// Filesystem contains everything needed to access the filesystem
type Filesystem struct {
	FS billy.Filesystem
}

func New(dir string) (*Filesystem, error) {
	i, err := os.Lstat(dir)
	if err != nil {
		return nil, err
	}
	if !i.IsDir() {
		return nil, errors.New("Not a directory")
	}

	ofs := osfs.New(dir)

	f := Filesystem{
		FS: ofs,
	}

	return &f, nil
}
