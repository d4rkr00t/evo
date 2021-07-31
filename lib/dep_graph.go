package lib

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

func (dg DepGraph) GetDependant(ws_name string) []string {
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
