package lib

import (
	"evo/main/lib/cache"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type WorkspacesMap struct {
	workspaces map[string]Workspace
	updated    map[string]bool
	dep_graph  DepGraph
	cache      *cache.Cache
}

func NewWorkspaceMap(root_path string, conf *Config, cc *cache.Cache) (WorkspacesMap, error) {
	var workspaces, err = get_workspaces(root_path, conf, cc)

	if err != nil {
		return WorkspacesMap{}, err
	}

	return WorkspacesMap{
		workspaces: workspaces,
		dep_graph:  NewDepGraph(&workspaces),
		updated:    map[string]bool{},
		cache:      cc,
	}, nil
}

func (wm *WorkspacesMap) Invalidate(targets []string) map[string]bool {
	var wg sync.WaitGroup
	var queue = make(chan []string)
	var mu sync.RWMutex

	wm.RehashAll()

	for name, ws := range wm.workspaces {
		wg.Add(1)
		go func(name string, ws Workspace) {
			var updated = false
			for _, target := range targets {
				var _, has_rule = ws.GetRule(target)
				if !has_rule {
					continue
				}
				var state_key = ClearTaskName(GetTaskName(target, ws.Name))
				if wm.cache.ReadData(state_key) != ws.hash {
					queue <- []string{name, "updated"}
					updated = true
					break
				}
			}

			// Not updated or no matched rules, ignore
			if !updated {
				queue <- []string{name}
			}
		}(name, ws)
	}

	go func() {
		for dat := range queue {
			mu.Lock()
			if len(dat) == 2 {
				wm.updated[dat[0]] = true
			}
			mu.Unlock()
			wg.Done()
		}
	}()

	wg.Wait()

	return wm.updated
}

func (wm *WorkspacesMap) RehashAll() {
	var visited = map[string]bool{}

	var process func(ws_name string)
	process = func(ws_name string) {
		var ws, ok = wm.workspaces[ws_name]
		if !ok {
			return
		}
		visited[ws_name] = true

		for dep := range ws.Deps {
			if _, ok := visited[dep]; !ok {
				process(dep)
			}
		}

		ws.Rehash(wm)
		wm.workspaces[ws.Name] = ws
	}

	for ws_name := range wm.workspaces {
		if _, ok := visited[ws_name]; !ok {
			process(ws_name)
		}
	}
}

func get_workspaces(root_path string, conf *Config, cc *cache.Cache) (map[string]Workspace, error) {
	var workspaces = make(map[string]Workspace)
	var wg sync.WaitGroup
	var queue = make(chan Workspace)
	var duplicates = []string{}

	for _, wc := range conf.Workspaces {
		var ws_glob = path.Join(root_path, wc, "package.json")
		var matches, _ = filepath.Glob(ws_glob)
		for _, ws_path := range matches {
			wg.Add(1)
			go func(ws_path string) {
				var excludes = conf.GetExcludes(ws_path)
				var rules = conf.GetAllRulesForWS(root_path, ws_path)
				queue <- NewWorkspace(root_path, ws_path, excludes, cc, rules)
			}(path.Dir(ws_path))
		}
	}

	go func() {
		for ws := range queue {
			if val, ok := workspaces[ws.Name]; ok {
				duplicates = append(duplicates, fmt.Sprintf("%s → %s → %s", ws.Name, ws.RelPath, val.RelPath))
			} else {
				workspaces[ws.Name] = ws
			}
			wg.Done()
		}
	}()

	wg.Wait()

	if len(duplicates) > 0 {
		return workspaces, fmt.Errorf("duplicate workspaces [ %s ]", strings.Join(duplicates, " | "))
	}

	return workspaces, nil
}
