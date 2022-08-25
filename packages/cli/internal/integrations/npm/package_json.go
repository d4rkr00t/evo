package npm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type PackageJson struct {
	Path             string
	Name             string
	Version          string
	Dependencies     map[string]string
	DevDependencies  map[string]string
	PeerDependencies map[string]string
	Bin              interface{}
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

func (pjs *PackageJson) GetBin() map[string]string {
	switch x := pjs.Bin.(type) {
	case string:
		var bins = map[string]string{}
		bins[pjs.Name] = x
		return bins
	case interface{}:
		return toStringString(x.(map[string]interface{}))
	default:
		return map[string]string{}
	}
}

func toStringString(mapInterface map[string]interface{}) map[string]string {
	var mapString = map[string]string{}

	for key, value := range mapInterface {
		strKey := fmt.Sprintf("%v", key)
		strValue := fmt.Sprintf("%v", value)
		mapString[strKey] = strValue
	}

	return mapString
}
