package fileutils

import (
	"path/filepath"
)

func RelativePathFileList(base string, flist []string) []string {
	for i, fpath := range flist {
		var relfpath, _ = filepath.Rel(base, fpath)
		flist[i] = relfpath
	}

	return flist
}
