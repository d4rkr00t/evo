package lib

import (
	"runtime"
	"strings"
	"sync"
)

type deps = map[string][]string

type DepGraph struct {
	direct  deps
	inverse deps
}

func NewDepGraph(workspaces *WorkspacesMap) DepGraph {
	var direct = build_direct(workspaces)
	var inverse = build_inverse(workspaces)
	return DepGraph{
		direct,
		inverse,
	}
}

func (dg DepGraph) HasCycles() (bool, string) {
	var visited = map[string]int{}

	var dfs func(ws_name string, path []string) (bool, []string)
	dfs = func(ws_name string, path []string) (bool, []string) {
		if visited[ws_name] == 2 {
			return false, path
		}

		if visited[ws_name] == 1 {
			return true, path
		}

		visited[ws_name] = 1
		for _, dep_name := range dg.direct[ws_name] {
			var cycle, path = dfs(dep_name, append(append([]string{}, path...), dep_name))
			if cycle {
				return true, path
			}
		}

		visited[ws_name] = 2
		return false, path
	}

	for ws_name := range dg.direct {
		var cycle, path = dfs(ws_name, []string{ws_name})
		if cycle {
			return true, strings.Join(path, " → ")
		}
	}

	return false, ""
}

func (dg DepGraph) GetAllDependant(ws_name string) []string {
	var dependant = map[string]bool{}
	var queue = dg.inverse[ws_name]

	for len(queue) > 0 {
		var cur = queue[0]
		queue = queue[1:]
		if _, ok := dependant[cur]; !ok {
			dependant[cur] = true
			queue = append(queue, dg.inverse[cur]...)
		}
	}

	var result = []string{}
	for key := range dependant {
		result = append(result, key)
	}
	return result
}

func (dg DepGraph) GetAffected(workspaces *WorkspacesMap, updated *map[string]string) map[string]string {
	var affected = map[string]string{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var guard = make(chan struct{}, runtime.NumCPU())

	for ws_name, ws_hash := range *updated {
		mu.Lock()
		affected[ws_name] = ws_hash
		mu.Unlock()
		for _, name := range dg.GetAllDependant(ws_name) {
			if _, ok := (*updated)[name]; !ok {
				wg.Add(1)
				guard <- struct{}{}
				go func(name string) {
					var hash = (*workspaces)[name].Hash(workspaces)
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

func build_direct(workspaces *WorkspacesMap) deps {
	var graph = make(deps)

	for _, ws := range *workspaces {
		if graph[ws.Name] == nil {
			graph[ws.Name] = []string{}
		}

		for name := range ws.Deps {
			if _, ok := (*workspaces)[name]; ok {
				graph[ws.Name] = append(graph[ws.Name], name)
			}
		}
	}

	return graph
}

func build_inverse(workspaces *WorkspacesMap) deps {
	var graph = make(deps)

	for _, ws := range *workspaces {

		for name := range ws.Deps {
			if _, ok := (*workspaces)[name]; ok {
				if graph[name] == nil {
					graph[name] = []string{}
				}
				graph[name] = append(graph[name], ws.Name)
			}
		}
	}

	return graph
}
