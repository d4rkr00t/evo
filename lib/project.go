package lib

import (
	"path"
	"path/filepath"
	"sync"
)

type Project struct {
	Cwd          string
	Package_json PackageJson
	Workspaces   map[string]Workspace
}

func NewProject(cwd string) Project {
	var package_json = NewPackageJson(cwd + "/package.json")
	var workspaces = get_workspaces_list(cwd, package_json.Workspaces)
	return Project{
		cwd,
		package_json,
		workspaces,
	}
}

func get_workspaces_list(cwd string, workspaces_config []string) map[string]Workspace {
	var workspaces = make(map[string]Workspace)
	var wg sync.WaitGroup

	for _, wc := range workspaces_config {
		var ws_glob = path.Join(cwd, wc, "package.json")
		var matches, _ = filepath.Glob(ws_glob)
		for _, ws_path := range matches {
			wg.Add(1)
			go func(ws_path string) {
				var workspace = NewWorkspace(cwd, path.Dir(ws_path))
				workspaces[workspace.Name] = workspace
				wg.Done()
			}(ws_path)
		}
	}

	wg.Wait()

	return workspaces
}
