package project

import (
	"path"
	"path/filepath"
	"scu/main/workspace"
	"sync"
)

type Project struct {
	cwd          string
	package_json PackageJson
	workspaces   []workspace.Workspace
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

func get_workspaces_list(cwd string, workspaces_config []string) []workspace.Workspace {
	var workspaces []workspace.Workspace
	var wg sync.WaitGroup

	for _, wc := range workspaces_config {
		var ws_glob = path.Join(cwd, wc, "package.json")
		var matches, _ = filepath.Glob(ws_glob)
		for _, ws_path := range matches {
			wg.Add(1)
			go func(ws_path string) {
				workspaces = append(workspaces, workspace.NewWorkspace(path.Dir(ws_path)))
				wg.Done()
			}(ws_path)
		}
	}

	wg.Wait()

	return workspaces
}
