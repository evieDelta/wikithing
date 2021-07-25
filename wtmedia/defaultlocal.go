package wtmedia

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	lru "github.com/hashicorp/golang-lru"
)

func NewDefaultLocal(dir string, cacheSize int) (*DefaultLocal, error) {
	fs := osfs.New(dir)
	ca, err := lru.NewARC(cacheSize)
	if err != nil {
		return nil, err
	}

	return &DefaultLocal{
		FS: fs,

		Cache: ca,
	}, nil
}

// DefaultLocal defines a media store that is a part of the same application
type DefaultLocal struct {
	FS billy.Filesystem

	Cache *lru.ARCCache
}

const metaExtension = ".json"

func (d *DefaultLocal) servBinary(hash string, meta ObjectMeta) (io.ReadCloser, string, error) {
	data, ok := d.getFromCache(hash)
	if ok {
		return io.NopCloser(
			bytes.NewReader(data),
		), meta.Type.Mime, nil
	}

	f, err := d.FS.Open(hash)

	return f, meta.Type.Mime, err
}

func (d *DefaultLocal) readMeta(hash string) (meta ObjectMeta, err error) {
	key := hash + metaExtension

	if d.Cache.Contains(key) {
		c, ok := d.Cache.Get(key)
		if !ok {
			// incredibly unlikely to happen since we just checked 0.0003 seconds ago but just in case
			goto oops
		}
		meta, ok = c.(ObjectMeta)
		if !ok {
			// once again this shouldn't happen, but just in case we might as well have a backup plan
			// also we remove it so it can put the proper one in place
			d.Cache.Remove(key)
			goto oops
		}
	}
oops:

	f, err := d.FS.Open(key)
	if err != nil {
		return meta, err
	}

	err = json.NewDecoder(f).Decode(&meta)
	if err != nil {
		return meta, err
	}

	d.Cache.Add(key, meta)

	return
}

func (d *DefaultLocal) saveMeta(hash string, meta ObjectMeta) (err error) {
	key := hash + metaExtension

	f, err := d.FS.OpenFile(key, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}

	var j *json.Encoder
	j = json.NewEncoder(f)
	j.SetIndent("", "	")
	j.SetEscapeHTML(false)
	return j.Encode(meta)
}

func (d *DefaultLocal) getFileFromEither(key string) (data []byte, err error) {
	data, ok := d.getFromCache(key)
	if ok {
		return data, nil
	}

	return d.getFromFile(key)
}

func (d *DefaultLocal) getFromFile(key string) (data []byte, err error) {
	f, err := d.FS.Open(key)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(f)
}

func (d *DefaultLocal) getFromCache(key string) (data []byte, ok bool) {
	if !d.Cache.Contains(key) {
		return
	}

	c, ok := d.Cache.Get(key)
	if !ok {
		// incredibly unlikely to happen since we just checked 0.0003 seconds ago but just in case
		return nil, false
	}

	data, ok = c.([]byte)
	if !ok {
		// once again this shouldn't happen, but just in case we might as well have a backup plan
		// also we remove it so it can put the proper one in place
		d.Cache.Remove(key)
		return nil, false
	}

	return data, true
}
