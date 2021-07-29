package lib

import (
	"fmt"
	"path"
	"path/filepath"
	"scu/main/lib/cache"
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

func (p Project) Invalidate(ws_list []string, cc cache.Cache) []string {
	var updated []string
	var is_all = len(ws_list) == 0
	var wg sync.WaitGroup
	var queue = make(chan string)

	if is_all {
		fmt.Println("Invalidating all packages!")
	}

	for name, ws := range p.Workspaces {
		wg.Add(1)
		go func(name string, ws Workspace) {
			var hash = ws.Hash()
			fmt.Println(name, hash)
			if !cc.Has(hash) {
				queue <- name
			} else {
				queue <- ""
			}
		}(name, ws)
	}

	go func() {
		for name := range queue {
			if len(name) > 0 {
				updated = append(updated, name)
			}
			wg.Done()
		}
	}()

	wg.Wait()

	return updated
}

func get_workspaces_list(cwd string, workspaces_config []string) map[string]Workspace {
	var workspaces = make(map[string]Workspace)
	var wg sync.WaitGroup
	var queue = make(chan Workspace)

	for _, wc := range workspaces_config {
		var ws_glob = path.Join(cwd, wc, "package.json")
		var matches, _ = filepath.Glob(ws_glob)
		for _, ws_path := range matches {
			wg.Add(1)
			go func(ws_path string) {
				queue <- NewWorkspace(cwd, path.Dir(ws_path))
			}(ws_path)
		}
	}

	go func() {
		for ws := range queue {
			workspaces[ws.Name] = ws
			wg.Done()
		}
	}()

	wg.Wait()

	return workspaces
}
