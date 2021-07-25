package wtfs

import (
	"os"
	"time"

	"git.lan/wikithing"
)

func (f *Filesystem) initialise(p wikithing.Path, pre string) error {
	err := f.FS.MkdirAll(p.Path(), 0664)
	if err != nil {
		return err
	}

	return f.initLog(p, pre)
}

func (f *Filesystem) initLog(p wikithing.Path, pre string) error {
	return f.saveJSON(p, pre, manLog, wikithing.LogFile{})
}

func (f *Filesystem) appendLog(p wikithing.Path, pre string, le wikithing.LogEntry) error {
	var log = new(wikithing.LogFile)

	return f.atomicUpdateJSON(p, pre, manLog, log, func() error {
		log.Entries = append(log.Entries, le)

		return nil
	})
}

func (f *Filesystem) loadFile(p wikithing.Path, pre string, dat interface{}) error {
	return f.loadJSON(p, pre, manCurrent, dat)
}

func (f *Filesystem) updateFile(p wikithing.Path, pre string, log wikithing.LogEntry, dat interface{}) error {
	unlock, err := f.getLock(p, pre)
	if err != nil {
		return err
	}
	defer unlock()

	when := time.Now().UTC()
	log.When = when

	if _, err := f.FS.Stat(pathSub(p, pre, manCurrent)); err != nil && os.IsNotExist(err) {
		log.Action = wikithing.LogActionCreate
	} else {
		log.Action = wikithing.LogActionEdit
		err = f.moveFile(p, pre, manCurrent, when.Format(TimeFormat))
		if err != nil {
			return err
		}
	}

	err = f.appendLog(p, pre, log)
	if err != nil {
		return err
	}

	return f.saveJSON(p, pre, manCurrent, dat)
}
