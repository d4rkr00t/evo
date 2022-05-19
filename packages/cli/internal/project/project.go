package project

import (
	gocontext "context"
	"evo/internal/context"
	"evo/internal/errors"
	"evo/internal/task_graph"
	"evo/internal/workspace"
	"fmt"
	"path"
	"strings"
	"sync"

	mapset "github.com/deckarep/golang-set"
	"github.com/pyr-sh/dag"
	"golang.org/x/sync/semaphore"
)

type ProjectWorkspacesMap = sync.Map

type Project struct {
	Path            string
	Config          ProjectConfig
	WorkspacesNames []string

	dependencyGraph dag.AcyclicGraph
	workspacesMap   ProjectWorkspacesMap
}

func NewProject(configPath string) (Project, error) {
	var project Project
	var projectConfig, err = LoadConfig(configPath)

	if err != nil {
		return project, errors.Wrap(errors.ErrorProjectConfigCouldntParse, fmt.Sprintf("Error in '%s'", ProjectConfigFileName), err)
	}

	var projectWorkspacesMap sync.Map
	var projectPath = path.Dir(configPath)
	var workspacesConfigsMap = DiscoverWorkspaces(projectPath, projectConfig.Workspaces)
	var duplicates = []string{}
	var workspacesNames = []string{}

	workspacesConfigsMap.Range(func(key, value any) bool {
		var wsPath = key.(string)
		var wsConfig = value.(workspace.WorkspaceConfig)
		var targets = GetAllTargetsForWorkspace(projectPath, &projectConfig, wsPath, &wsConfig)
		var excludes = append(append([]string{}, projectConfig.GetExcludes(projectPath, wsPath)...), wsConfig.Excludes...)

		var _, ok = projectWorkspacesMap.Load(wsConfig.Name)

		if ok {
			duplicates = append(duplicates, wsConfig.Name)
		} else {
			projectWorkspacesMap.Store(wsConfig.Name, workspace.New(wsConfig.Name, wsPath, targets, excludes))
			workspacesNames = append(workspacesNames, wsConfig.Name)
		}

		return true
	})

	if len(duplicates) > 0 {
		return project, errors.New(errors.ErrorProjectDuplicateWorkspaces, fmt.Sprintf("duplicate workspaces [ %s ]", strings.Join(duplicates, " | ")))
	}

	project = Project{
		Path:            projectPath,
		Config:          projectConfig,
		WorkspacesNames: workspacesNames,

		workspacesMap: projectWorkspacesMap,
	}

	return project, nil
}

func (p *Project) Size() int {
	return len(p.WorkspacesNames)
}

func (p *Project) Load(wsName string) (*workspace.Workspace, bool) {
	var value, ok = p.workspacesMap.Load(wsName)
	var ws workspace.Workspace

	if ok {
		ws = value.(workspace.Workspace)
	}

	return &ws, ok
}

func (p *Project) Store(ws *workspace.Workspace) {
	p.workspacesMap.Store(ws.Name, *ws)
}

func (p *Project) BuildDependencyGraph() {
	p.workspacesMap.Range(func(key, value any) bool {
		p.dependencyGraph.Add(key.(string))
		return true
	})

	p.workspacesMap.Range(func(key, value any) bool {
		var ws = value.(workspace.Workspace)
		for _, dep := range ws.Deps {
			if dep.Type == "local" {
				p.dependencyGraph.Connect(dag.BasicEdge(ws.Name, dep.Name))
			}
		}
		return true
	})
}

func (p *Project) Walk(fn func(ws *workspace.Workspace) error, concurency int) {
	var cc = gocontext.TODO()
	var sem = semaphore.NewWeighted(int64(concurency))

	p.dependencyGraph.Walk(func(v dag.Vertex) error {
		var wsName = fmt.Sprint(v)
		if err := sem.Acquire(cc, 1); err != nil {
			panic(fmt.Sprintf("Failed to acquire semaphore: %v", err))
		}
		defer sem.Release(1)
		var ws, _ = p.Load(wsName)
		return fn(ws)
	})
}

func (p *Project) Validate() error {
	var cycles = p.dependencyGraph.Cycles()

	if len(cycles) == 0 {
		return nil
	}

	var msg = []string{"cycles in the dependency graph:"}
	for _, cycle := range cycles {
		msg = append(msg, fmt.Sprintf("– %s", cycle))
	}

	return errors.New(errors.ErrorProjectDepGraphCycle, strings.Join(msg, "\n"))
}

func (p *Project) RehashAllWorkspaces(ctx *context.Context) {
	p.Walk(func(ws *workspace.Workspace) error {
		defer ctx.Tracer.Event(fmt.Sprintf("invalidating workspace %s", ws.Name)).Done()
		ws.Rehash(&p.workspacesMap)
		p.Store(ws)
		return nil
	}, ctx.Concurrency)
}

func (p *Project) ReduceToScope(scope []string) {
	var workspacesInScope sync.Map
	var visited = mapset.NewSet()
	var workspacesNames = []string{}

	var idx = 0
	for idx < len(scope) {
		var scopeName = scope[idx]
		idx += 1

		var ws, ok = p.Load(scopeName)
		if !ok || visited.Contains(ws.Name) {
			continue
		}

		visited.Add(ws.Name)
		workspacesInScope.Store(ws.Name, *ws)
		workspacesNames = append(workspacesNames, ws.Name)

		for _, wsDep := range ws.Deps {
			if wsDep.Type == "local" && !visited.Contains(wsDep.Name) {
				scope = append(scope, wsDep.Name)
			}
		}
	}

	p.workspacesMap = workspacesInScope
	p.WorkspacesNames = workspacesNames
}

func (p *Project) GetAffectedWorkspaces(ctx *context.Context, targetsNames []string) []string {
	var affectedWorkspaces = []string{}
	var mx sync.Mutex
	p.Walk(func(ws *workspace.Workspace) error {
		for _, targetName := range targetsNames {
			if target, ok := ws.Targets[targetName]; ok {
				var task = task_graph.NewTask(ws, targetName, &target)
				if task.Invalidate(&ctx.Cache) {
					mx.Lock()
					affectedWorkspaces = append(affectedWorkspaces, ws.Name)
					mx.Unlock()
					break
				}
			}
		}
		return nil
	}, ctx.Concurrency)
	return affectedWorkspaces
}

func (p *Project) GetWorkspacesMatchingFiles(filesList []string) []string {
	var workspacesNames = []string{}

	for _, wsName := range p.WorkspacesNames {
		var ws, _ = p.Load(wsName)

		for _, file := range filesList {
			if strings.HasPrefix(file, ws.Path) {
				workspacesNames = append(workspacesNames, wsName)
				break
			}
		}
	}

	return workspacesNames
}
