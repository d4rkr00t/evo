package npm

import (
	"encoding/json"
	"io/ioutil"
)

type PackageJson struct {
	Path             string
	Name             string
	Version          string
	Dependencies     map[string]string
	DevDependencies  map[string]string
	PeerDependencies map[string]string
	Bin              map[string]string
}

func NewPackageJson(packageJsonPath string) (PackageJson, error) {
	var p PackageJson
	var dat, err = ioutil.ReadFile(packageJsonPath)
	if err != nil {
		return p, err
	}

	err = json.Unmarshal(dat, &p)
	if err != nil {
		return p, err
	}

	p.Path = packageJsonPath
	return p, nil
}

func (pjs *PackageJson) GetAllDependencies() map[string]string {
	var allDeps = map[string]string{}

	for name, version := range pjs.DevDependencies {
		allDeps[name] = version
	}

	for name, version := range pjs.Dependencies {
		allDeps[name] = version
	}

	return allDeps
}
