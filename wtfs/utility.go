package wtfs

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path"
	"time"

	"git.lan/wikithing"
)

// TimeFormat is the format used in past version timestamps
const TimeFormat = "2006-01-02_15:04:05"

const (
	manLock    = ".lockfile"
	manCurrent = "current"
	manLog     = "log"
)

func pathLock(p wikithing.Path, pre string) string    { return pathSub(p, pre, manLock) }
func pathCurrent(p wikithing.Path, pre string) string { return pathSub(p, pre, manCurrent) }
func pathLog(p wikithing.Path, pre string) string     { return pathSub(p, pre, manLog) }
func pathWhen(p wikithing.Path, x string, at time.Time) string {
	return pathSub(p, x, at.Format(TimeFormat))
}
func pathSub(p wikithing.Path, prefix, sub string) string {
	return path.Join(prefix, p.Path(), sub+Extension)
}

func (f *Filesystem) saveJSON(p wikithing.Path, pre, sub string, dat interface{}) error {
	unlock, err := f.getLock(p, pre)
	if err != nil {
		return err
	}
	defer unlock()

	d, err := f.FS.OpenFile(pathSub(p, pre, sub), os.O_RDWR|os.O_CREATE, 0664)
	defer d.Close()
	if err != nil {
		return err
	}

	var j *json.Encoder
	j = json.NewEncoder(d)
	j.SetIndent("", "	")
	j.SetEscapeHTML(false)
	return j.Encode(dat)
}

func (f *Filesystem) loadJSON(p wikithing.Path, pre, sub string, dat interface{}) error {
	unlock, err := f.getLock(p, pre)
	if err != nil {
		return err
	}
	defer unlock()

	d, err := f.FS.OpenFile(pathSub(p, pre, sub), os.O_RDWR|os.O_CREATE, 0664)
	defer d.Close()
	if err != nil {
		return err
	}

	return json.NewDecoder(d).Decode(dat)
}

func (f *Filesystem) atomicUpdateJSON(p wikithing.Path, pre, sub string,
	/**/ dat interface{}, fn func() error) error {

	d, err := f.FS.OpenFile(pathSub(p, pre, sub), os.O_RDWR|os.O_CREATE, 0664)
	defer d.Close()
	if err != nil {
		return err
	}
	err = d.Lock()
	if err != nil {
		return err
	}
	defer func() {
		err := d.Unlock()
		if err != nil {
			// this better not ever happen
			log.Println("!!!UH OH BIG ERROR. FAILED TO RELEASE FILE LOCK ON: `", pathSub(p, pre, sub), "` BECAUSE:", err)
		}
	}()

	err = json.NewDecoder(d).Decode(dat)
	if err != nil {
		return err
	}
	err = fn()
	if err != nil {
		return err
	}
	err = d.Truncate(0)
	if err != nil {
		return err
	}

	var j *json.Encoder
	j = json.NewEncoder(d)
	j.SetIndent("", "	")
	j.SetEscapeHTML(false)
	return j.Encode(dat)
}

func (f *Filesystem) moveFile(p wikithing.Path, pre string, subfrom, subto string) error {
	return f.FS.Rename(pathSub(p, pre, subfrom), pathSub(p, pre, subto))
}

func (f *Filesystem) copyFile(p wikithing.Path, pre string, subfrom, subto string) error {
	d, err := f.FS.OpenFile(pathSub(p, pre, subfrom), os.O_RDONLY, 0664)
	defer d.Close()
	if err != nil {
		return err
	}

	t, err := f.FS.OpenFile(pathSub(p, pre, subto), os.O_RDWR|os.O_CREATE, 0664)
	defer t.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(t, d)
	return err
}

// getLock fetches the lock file for a managed data group preventing other concurrent processes from also modifying it
func (f *Filesystem) getLock(p wikithing.Path, prefix string) (func(), error) {
	file := pathLock(p, prefix)
	l, err := f.FS.OpenFile(file, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		l.Close()
		return nil, err
	}

	return func() {
		l.Close()
		f.FS.Remove(file)
	}, nil
}
