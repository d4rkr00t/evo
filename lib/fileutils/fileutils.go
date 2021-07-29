package fileutils

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
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
	type Pair struct {
		fpath string
		data  []byte
	}

	var h = sha1.New()
	var wg sync.WaitGroup
	var queue = make(chan Pair)

	for _, fpath := range flist {
		wg.Add(1)
		go func(fpath string) {
			var data, _ = ioutil.ReadFile(fpath)
			queue <- Pair{fpath, data}
		}(fpath)
	}

	var files []Pair

	go func() {
		for data := range queue {
			files = append(files, data)
			wg.Done()
		}
	}()

	wg.Wait()

	sort.Slice(files[:], func(i, j int) bool {
		return files[i].fpath > files[j].fpath
	})

	for _, pair := range files {
		h.Write(pair.data)
	}

	return hex.EncodeToString(h.Sum(nil))
}

func RelativePathFileList(base string, flist []string) []string {
	for i, fpath := range flist {
		var relfpath, _ = filepath.Rel(base, fpath)
		flist[i] = relfpath
	}

	return flist
}
