package fileutils

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

func GetFileHash(fpath string) string {
	var f, _ = os.Open(fpath)
	defer f.Close()

	var h = sha1.New()
	io.Copy(h, f)

	return hex.EncodeToString(h.Sum(nil))
}
