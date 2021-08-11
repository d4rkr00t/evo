package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"scu/main/lib/cache"
	"scu/main/lib/fileutils"
)

type PackageJson struct {
	Path         string
	Name         string
	Version      string
	Scu          Config
	Dependencies map[string]string
}

func NewPackageJson(path string) PackageJson {
	var p PackageJson
	var dat, _ = ioutil.ReadFile(path)
	json.Unmarshal(dat, &p)
	p.Path = path
	return p
}

func (p PackageJson) GetConfig() Config {
	return p.Scu
}

func (p PackageJson) Invalidate(cc *cache.Cache) bool {
	var hash = p.GetHash()
	return cc.ReadData(p.GetStateKey()) != hash
}

func (p PackageJson) GetStateKey() string {
	return fmt.Sprintf("%s-packagejson", p.Name)
}

func (p PackageJson) GetHash() string {
	return fileutils.GetFileHash(p.Path)
}

func (p PackageJson) CacheState(c *cache.Cache) {
	c.CacheData(p.GetStateKey(), p.GetHash())
}
