package lib

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"scu/main/lib/cache"
	"sync"
)

type WorkspacesMap = map[string]Workspace

type Project struct {
	Cwd          string
	Package_json PackageJson
	Workspaces   WorkspacesMap
	DepGraph     DepGraph
}

func NewProject(cwd string) Project {
	var package_json = NewPackageJson(path.Join(cwd, "package.json"))
	var workspaces = get_workspaces_list(cwd, &package_json.Scu)
	var dep_graph = NewDepGraph(&workspaces)

	return Project{
		Cwd:          cwd,
		Package_json: package_json,
		Workspaces:   workspaces,
		DepGraph:     dep_graph,
	}
}

func (p Project) Invalidate(ws_list []string, cmd string, cc *cache.Cache) map[string]string {
	var updated = map[string]string{}
	var is_all = len(ws_list) == 0
	var wg sync.WaitGroup
	var queue = make(chan []string)

	if is_all {
		fmt.Println("Invalidating all packages!")
	}

	for name, ws := range p.Workspaces {
		wg.Add(1)
		go func(name string, ws Workspace) {
			var updated, ws_hash = ws.Invalidate(cmd, cc)
			if updated {
				queue <- []string{name, ws_hash}
			} else {
				queue <- []string{}
			}
		}(name, ws)
	}

	go func() {
		for dat := range queue {
			if len(dat) > 0 {
				updated[dat[0]] = dat[1]
			}
			wg.Done()
		}
	}()

	wg.Wait()

	return updated
}

func (p Project) GetWs(name string) Workspace {
	return p.Workspaces[name]
}

func (p Project) GetDependant(ws_name string) map[string]string {
	var affected = map[string]string{}
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, name := range p.DepGraph.GetDependant(ws_name) {
		wg.Add(1)
		go func(name string) {
			mu.Lock()
			affected[name] = p.GetWs(name).Hash()
			mu.Unlock()
			wg.Done()
		}(name)
	}

	wg.Wait()

	return affected
}

func (p Project) GetAffected(workspaces *map[string]string) map[string]string {
	var affected = map[string]string{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var guard = make(chan struct{}, runtime.NumCPU())

	for ws_name, ws_hash := range *workspaces {
		mu.Lock()
		affected[ws_name] = ws_hash
		mu.Unlock()
		for _, name := range p.DepGraph.GetDependant(ws_name) {
			if _, ok := (*workspaces)[name]; !ok {
				wg.Add(1)
				guard <- struct{}{}
				go func(name string) {
					var hash = p.GetWs(name).Hash()
					mu.Lock()
					affected[name] = hash
					<-guard
					mu.Unlock()
					wg.Done()
				}(name)
			}
		}
	}

	wg.Wait()

	return affected
}

func (p Project) GetNodeModulesBinPath() string {
	return path.Join(p.Cwd, "node_modules", ".bin")
}

func (p Project) GetRule(name string, ws_path string) Rule {
	return p.Package_json.Scu.GetRule(name, ws_path)
}

func get_workspaces_list(cwd string, conf *Config) map[string]Workspace {
	var workspaces = make(map[string]Workspace)
	var wg sync.WaitGroup
	var queue = make(chan Workspace)

	for _, wc := range conf.Workspaces {
		var ws_glob = path.Join(cwd, wc, "package.json")
		var matches, _ = filepath.Glob(ws_glob)
		for _, ws_path := range matches {
			wg.Add(1)
			go func(ws_path string) {
				var includes, excludes = conf.GetInputs(path.Dir(ws_path))
				queue <- NewWorkspace(cwd, path.Dir(ws_path), includes, excludes)
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
