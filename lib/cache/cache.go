package cache

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/otiai10/copy"
)

type Cache struct {
	path string
}

type CacheDirIgnores = map[string]bool

func NewCache(project_path string) Cache {
	var folder = path.Join(project_path, ".cache")
	os.MkdirAll(folder, 0700)
	return Cache{path: folder}
}

func (c Cache) Has(key string) bool {
	var _, err = os.Lstat(c.get_cache_path(key))
	return err == nil
}

func (c Cache) CacheData(key string, data string) {
	var p = c.get_cache_path(key)
	ioutil.WriteFile(p, []byte(data), 0644)
}

func (c Cache) ReadData(key string) string {
	if !c.Has(key) {
		return ""
	}

	var dat, _ = ioutil.ReadFile(c.get_cache_path(key))
	return string(dat)
}

func (c Cache) CacheDir(key string, dpath string, ignores CacheDirIgnores) {
	copy.Copy(dpath, c.get_cache_path(key), copy.Options{
		Skip: func(src string) (bool, error) {
			var rel_src, _ = filepath.Rel(dpath, src)
			return ignores[rel_src], nil
		},
	})
}

func (c Cache) get_cache_path(p string) string {
	return path.Join(c.path, p)
}
