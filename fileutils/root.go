package fileutils

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

func GetFileHash(fpath string) string {
	var f, _ = os.Open(path.Join(fpath, "package.json"))
	defer f.Close()

	var h = sha1.New()
	io.Copy(h, f)

	return hex.EncodeToString(h.Sum(nil))
}

func GetFileListHash(flist []string) string {
	var h = sha1.New()
	var wg sync.WaitGroup
	var queue = make(chan []byte)

	for _, fpath := range flist {
		wg.Add(1)
		go func(fpath string) {
			var data, _ = ioutil.ReadFile(fpath)
			queue <- data
		}(fpath)
	}

	go func() {
		for data := range queue {
			h.Write(data)
			wg.Done()
		}
	}()

	wg.Wait()

	return hex.EncodeToString(h.Sum(nil))
}
