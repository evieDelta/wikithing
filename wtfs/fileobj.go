package wtfs

import "git.lan/wikithing"

// Media is the folder Media metadata goes in
const Media = "media"

func (f *Filesystem) LoadMedia(loc wikithing.Path) (a wikithing.FileObject, err error) {
	return a, f.loadFile(loc, Media, &a.FileConfig)
}

func (f *Filesystem) SaveMedia(loc wikithing.Path, media wikithing.FileObject, why wikithing.LogEntry) error {
	return f.updateFile(loc, Media, why, media)
}
