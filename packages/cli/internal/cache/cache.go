package cache

import (
	"io/ioutil"
	"os"
	"path"
)

type Cache struct {
	dirPath string
}

type CacheDirIgnores = map[string]bool

const DefaultCacheLocation = ".evo_cache"

func New(rootPath string, cacheDirName string) Cache {
	var dirPath = path.Join(rootPath, cacheDirName)
	return Cache{
		dirPath: dirPath,
	}
}

func (cache *Cache) Setup() {
	os.MkdirAll(cache.dirPath, 0700)
}

func (c *Cache) Has(key string) bool {
	var _, err = os.Lstat(c.GetCachePath(key))
	return err == nil
}

func (c Cache) CacheData(key string, data string) {
	var p = c.GetCachePath(key)
	ioutil.WriteFile(p, []byte(data), 0644)
}

func (c Cache) ReadData(key string) string {
	if !c.Has(key) {
		return ""
	}

	var dat, _ = ioutil.ReadFile(c.GetCachePath(key))
	return string(dat)
}

func (c Cache) GetCachePath(p string) string {
	return path.Join(c.dirPath, p)
}
