package lib

import (
	"path"
	"path/filepath"
)

type Workspace struct {
	Name string
	Path string
	Deps map[string]string
}

func NewWorkspace(project_path string, ws_path string) Workspace {
	var rel_ws_path, _ = filepath.Rel(project_path, ws_path)
	var package_json_path = path.Join(ws_path, "package.json")
	var package_json = NewPackageJson(package_json_path)

	return Workspace{
		Name: package_json.Name,
		Path: rel_ws_path,
		Deps: package_json.Dependencies,
	}
}
