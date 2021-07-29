package cache

import (
	"io/ioutil"
	"os"
	"path"
)

type Cache struct {
	path string
}

func NewCache(project_path string) Cache {
	var folder = path.Join(project_path, ".cache")
	os.MkdirAll(folder, 0700)
	return Cache{path: folder}
}

func (c Cache) Has(key string) bool {
	var _, err = os.Lstat(c.get_cache_path(key))
	return err != nil
}

func (c Cache) get_cache_path(p string) string {
	return path.Join(c.path, p)
}

func (c Cache) CacheFile(name string, filep string) {
	var dat, _ = ioutil.ReadFile(filep)
	var cpath = c.get_cache_path(name)
	ioutil.WriteFile(cpath, dat, 0644)
}

func (c Cache) CacheData(name string, dat []byte) {
	var cpath = c.get_cache_path(name)
	ioutil.WriteFile(cpath, dat, 0644)
}
