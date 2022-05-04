package lib

import (
	"context"
	"evo/main/lib/cache"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"evo/main/lib/goccm"

	mapset "github.com/deckarep/golang-set"
	"github.com/pyr-sh/dag"
	"golang.org/x/sync/semaphore"
)

type WorkspacesMap struct {
	workspaces sync.Map
	updated    mapset.Set
	dep_graph  dag.AcyclicGraph
	cache      *cache.Cache
	length     int32
}

func NewWorkspaceMap(root_path string, conf *Config, cc *cache.Cache) (WorkspacesMap, error) {
	var workspaces, num, err = get_workspaces(root_path, conf, cc)

	if err != nil {
		return WorkspacesMap{}, err
	}

	return WorkspacesMap{
		workspaces: workspaces,
		dep_graph:  NewDAGFromWorkspaces(&workspaces),
		updated:    mapset.NewSet(),
		cache:      cc,
		length:     num,
	}, nil
}

func (wm *WorkspacesMap) Load(ws_name string) (Workspace, bool) {
	var value, ok = wm.workspaces.Load(ws_name)
	if ok {
		return value.(Workspace), ok
	}
	var ws Workspace
	return ws, ok
}

func (wm *WorkspacesMap) Store(ws Workspace) {
	wm.workspaces.Store(ws.Name, ws)
}

func (wm *WorkspacesMap) Range(f func(key string, value Workspace) bool) {
	wm.workspaces.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(Workspace))
	})
}

func (wm *WorkspacesMap) Invalidate(ctx *Context) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	wm.RehashAll(ctx)

	wm.Range(func(_ string, ws Workspace) bool {
		wg.Add(1)

		go func(name string, ws Workspace) {
			for _, target := range ctx.target {
				var _, has_rule = ws.GetRule(target)
				if !has_rule {
					continue
				}
				var state_key = ClearTaskName(GetTaskName(target, ws.Name))
				if len(ctx.changed_files) > 0 {
					for _, file := range ctx.changed_files {
						if strings.HasPrefix(file, ws.RelPath) {
							mu.Lock()
							wm.updated.Add(ws.Name)
							mu.Unlock()
							break
						}
					}
				} else if wm.cache.ReadData(state_key) != ws.hash {
					mu.Lock()
					wm.updated.Add(ws.Name)
					mu.Unlock()
					break
				}
			}
			wg.Done()
		}(ws.Name, ws)

		return true
	})

	wg.Wait()
}

func (wm *WorkspacesMap) ReduceToScope(scope []string) {
	var all_in_scope sync.Map
	var visited = mapset.NewSet()
	var num int32 = 0

	var idx = 0
	for len(scope) > idx {
		var scope_name = scope[idx]
		if ws, ok := wm.Load(scope_name); ok {
			if _, ok = all_in_scope.Load(scope_name); !ok {
				all_in_scope.Store(scope_name, ws)
				num += 1
			}

			for dep_name := range ws.Deps {
				if _, ok := wm.Load(dep_name); ok {
					if _, ok := all_in_scope.Load(dep_name); !ok {
						scope = append(scope, dep_name)
					}
				}
			}

			for _, edge := range wm.dep_graph.EdgesFrom(ws.Name) {
				var tgt = fmt.Sprint(edge.Target())
				if visited.Contains(tgt) {
					continue
				}
				visited.Add(tgt)
				scope = append(scope, tgt)
			}
		}

		idx += 1
	}

	wm.workspaces = all_in_scope
	wm.dep_graph = NewDAGFromWorkspaces(&all_in_scope)
	wm.length = num
}

func (wm *WorkspacesMap) RehashAll(ctx *Context) {
	var cc = context.TODO()
	var sem = semaphore.NewWeighted(int64(ctx.concurrency))

	wm.dep_graph.Walk(func(vx dag.Vertex) error {
		var ws_name = fmt.Sprint(vx)
		if err := sem.Acquire(cc, 1); err != nil {
			panic(fmt.Sprintf("Failed to acquire semaphore: %v", err))
		}
		defer sem.Release(1)

		var ws, _ = wm.Load(ws_name)
		ws.Rehash(wm)
		wm.Store(ws)

		return nil
	})
}

func (wm *WorkspacesMap) Validate() error {
	var cycles = wm.dep_graph.Cycles()
	if len(cycles) > 0 {
		return fmt.Errorf("cycle in the dependecy graph [ %s ]", cycles[0])
	}
	return nil
}

func get_workspaces(root_path string, conf *Config, cc *cache.Cache) (sync.Map, int32, error) {
	var workspaces sync.Map
	var duplicates = []string{}
	var ccm = goccm.New(runtime.NumCPU())
	var num int32 = 0

	for _, wc := range conf.Workspaces {
		var ws_glob = path.Join(root_path, wc, "package.json")
		var matches, _ = filepath.Glob(ws_glob)
		for _, ws_path := range matches {
			ccm.Wait()
			go func(ws_path string) {
				defer ccm.Done()
				var excludes = conf.GetExcludes(ws_path)
				var rules = conf.GetAllRulesForWS(root_path, ws_path)
				var ws = NewWorkspace(root_path, ws_path, excludes, cc, rules)
				if val, ok := workspaces.Load(ws.Name); ok {
					duplicates = append(duplicates, fmt.Sprintf("%s → %s → %s", ws.Name, ws.RelPath, val.(Workspace).RelPath))
				} else {
					workspaces.Store(ws.Name, ws)
					atomic.AddInt32(&num, 1)
				}
			}(path.Dir(ws_path))
		}
	}

	ccm.WaitAllDone()

	if len(duplicates) > 0 {
		return workspaces, num, fmt.Errorf("duplicate workspaces [ %s ]", strings.Join(duplicates, " | "))
	}

	return workspaces, num, nil
}
