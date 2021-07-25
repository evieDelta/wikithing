package wikithing

import "git.lan/wikithing/etc/sid"

type FileClass uint

// FileClassTypes
const (
	FileClassUndefined = 0
	FileClassOther     = 1
	FileClassImage     = 2
	FileClassAudio     = 3
	FileClassVideo     = 4
)

// FileObject is an representation of a single media file
type FileObject struct {
	ID sid.ID

	FileConfig
}

type FileConfig struct {
	Class FileClass // the "kind" of a file, eg audio, video, image
	Mime  string    // should be a mime-type

	Hash string // the unique hash to a file, allows updating a file object
}
