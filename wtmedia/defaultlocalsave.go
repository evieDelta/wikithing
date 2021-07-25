package wtmedia

import (
	"os"
	"time"
)

func (d *DefaultLocal) Put(hash string, kind TypeMeta, data []byte) error {
	meta := ObjectMeta{
		Type:    kind,
		Created: time.Now().UTC(),
	}

	err := d.saveMeta(hash, meta)
	if err != nil {
		return err
	}

	f, err := d.FS.OpenFile(hash, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0664)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	return err
}

func (d *DefaultLocal) Rem(hash string) error {
	err := d.FS.Remove(hash)
	if err != nil {
		return err
	}
	return d.FS.Remove(hash + metaExtension)
}
