package fileutils

import (
	"os"
)

func Exist(fpath string) bool {
	var _, err = os.Stat(fpath)
	return !os.IsNotExist(err)
}
