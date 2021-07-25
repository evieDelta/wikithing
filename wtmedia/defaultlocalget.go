package wtmedia

import "io"

// Get ...
func (d *DefaultLocal) Get(hash string, q QueryData) (io.ReadCloser, string, error) {
	fmeta, err := d.readMeta(hash)
	if err != nil {
		return nil, "", err
	}

	switch fmeta.Type.Kind {
	case KindImage:
		return d.servImage(hash, fmeta, q)
	default:
		// if its an unrecognised value just use the default binary getter
		fallthrough
	case KindBinary:
		return d.servBinary(hash, fmeta)
	}
}

func (d *DefaultLocal) GetMeta(hash string) (meta ObjectMeta, err error) {
	return d.readMeta(hash)
}
