package lib

import (
	"encoding/json"
	"evo/main/lib/cache"
	"evo/main/lib/fileutils"
	"fmt"
	"io/ioutil"
)

type PackageJson struct {
	Path            string
	Name            string
	Version         string
	Evo             Config
	Dependencies    map[string]string
	DevDependencies map[string]string
	Bin             map[string]string
}

func NewPackageJson(package_json_path string) (PackageJson, error) {
	var p PackageJson
	var dat, err = ioutil.ReadFile(package_json_path)
	if err != nil {
		return p, err
	}
	json.Unmarshal(dat, &p)
	p.Path = package_json_path
	return p, nil
}

func (p PackageJson) GetConfig() Config {
	return p.Evo
}

func (p PackageJson) Invalidate(cc *cache.Cache) bool {
	var hash = p.GetHash()
	return cc.ReadData(p.GetStateKey()) != hash
}

func (p PackageJson) GetStateKey() string {
	return fmt.Sprintf("%s-packagejson", ClearTaskName(p.Name))
}

func (p PackageJson) GetHash() string {
	return fileutils.GetFileHash(p.Path)
}

func (p PackageJson) CacheState(c *cache.Cache) {
	println(p.GetStateKey(), p.GetHash())
	c.CacheData(p.GetStateKey(), p.GetHash())
}
