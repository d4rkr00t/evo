package fileutils

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"path"
)

func GetFileHash(fpath string) string {
	var f, _ = os.Open(path.Join(fpath, "package.json"))
	defer f.Close()

	var h = sha1.New()
	io.Copy(h, f)

	return hex.EncodeToString(h.Sum(nil))
}
