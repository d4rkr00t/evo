package project

import (
	"evo/internal/ccm"
	"evo/internal/context"
	"evo/internal/errors"
	"evo/internal/workspace"
	"fmt"
	"path"
	"strings"
	"sync"

	mapset "github.com/deckarep/golang-set"
)

type ProjectWorkspacesMap = sync.Map

type Project struct {
	Path            string
	Config          ProjectConfig
	WorkspacesNames []string
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
			var ws = workspace.New(wsConfig.Name, wsPath, targets, excludes)
			projectWorkspacesMap.Store(wsConfig.Name, ws)
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
	var ws *workspace.Workspace

	if ok {
		ws = value.(*workspace.Workspace)
	}

	return ws, ok
}

func (p *Project) Store(ws *workspace.Workspace) {
	p.workspacesMap.Store(ws.Name, ws)
}

func (p *Project) RehashAllWorkspaces(ctx *context.Context) {
	var cm = ccm.New(ctx.Concurrency)

	p.Range(func(ws *workspace.Workspace) bool {
		cm.Add()
		go func(ws *workspace.Workspace) {
			defer ctx.Tracer.Event(fmt.Sprintf("invalidating workspace %s", ws.Name)).Done()
			ws.Rehash(&p.workspacesMap)
			cm.Done()
		}(ws)
		return true
	})

	cm.Wait()
}

func (p *Project) Range(fn func(ws *workspace.Workspace) bool) {
	p.workspacesMap.Range(func(_, value any) bool {
		var ws = value.(*workspace.Workspace)
		return fn(ws)
	})
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
		workspacesInScope.Store(ws.Name, ws)
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
