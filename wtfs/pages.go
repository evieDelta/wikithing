package wtfs

import (
	"git.lan/wikithing"
)

// Pages is the folder Pages go in
const Pages = "pages"

func (f *Filesystem) LoadPage(loc wikithing.Path) (a wikithing.Article, err error) {
	return a, f.loadFile(loc, Pages, &a)
}

func (f *Filesystem) SavePage(loc wikithing.Path, page wikithing.Article, why wikithing.LogEntry) error {
	return f.updateFile(loc, Pages, why, page)
}
