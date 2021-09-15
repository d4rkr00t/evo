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
	hashes     map[string]string
	updated    map[string]bool
	affected   map[string]bool
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
		hashes:     map[string]string{},
		updated:    map[string]bool{},
		affected:   map[string]bool{},
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
			mu.RLock()
			var ws_hash = wm.hashes[ws.Name]
			mu.RUnlock()
			var updated = false
			for _, target := range targets {
				var _, has_rule = ws.GetRule(target)
				if !has_rule {
					continue
				}
				var state_key = ClearTaskName(GetTaskName(target, ws.Name))
				if wm.cache.ReadData(state_key) != ws_hash {
					queue <- []string{name, ws_hash, "updated"}
					updated = true
					break
				}
			}
			if !updated {
				if ws.GetCacheState() != ws_hash {
					queue <- []string{name, ws_hash, "updated"}
				} else {
					queue <- []string{name, ws_hash}
				}
			}
		}(name, ws)
	}

	go func() {
		for dat := range queue {
			mu.Lock()
			if len(dat) == 3 {
				wm.updated[dat[0]] = len(dat) == 3
			}
			wm.hashes[dat[0]] = dat[1]
			mu.Unlock()
			wg.Done()
		}
	}()

	wg.Wait()

	return wm.updated
}

func (wm *WorkspacesMap) GetAffected() map[string]bool {
	for ws_name := range wm.updated {
		wm.affected[ws_name] = true
		for _, name := range wm.dep_graph.GetAllDependant(ws_name) {
			if _, ok := (wm.updated)[name]; !ok {
				wm.affected[name] = true
			}
		}
	}

	return wm.affected
}

func (wm *WorkspacesMap) RehashAll() {
	var visited = map[string]bool{}

	var process func(ws_name string)
	process = func(ws_name string) {
		var ws = wm.workspaces[ws_name]
		visited[ws_name] = true

		for dep := range ws.Deps {
			if _, ok := visited[dep]; !ok {
				process(dep)
			}
		}

		wm.hashes[ws_name] = ws.Hash(wm)
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
