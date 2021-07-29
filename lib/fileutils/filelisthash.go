package fileutils

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/ioutil"
)

func GetFileListHash(flist []string) string {
	var h = sha1.New()

	for _, fpath := range flist {
		var data, err = ioutil.ReadFile(fpath)
		if err == nil {
			io.WriteString(h, string(data))
		}
	}

	return hex.EncodeToString(h.Sum(nil))
}
