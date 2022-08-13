package project

import (
	"evo/internal/ccm"
	"evo/internal/workspace"
	"path"
	"path/filepath"
	"runtime"
	"sync"
)

func DiscoverWorkspaces(rootPath string, workspacesGlobsList []string) sync.Map {
	var workspacesConfigsMap sync.Map
	if len(workspacesGlobsList) == 0 {
		return workspacesConfigsMap
	}

	// TODO: use concurency settings
	var cm = ccm.New(runtime.NumCPU())

	for _, wc := range workspacesGlobsList {
		var wsGlob = path.Join(rootPath, wc, workspace.WorkspaceConfigFileName)
		var matches, _ = filepath.Glob(wsGlob)

		for _, wsConfigPath := range matches {
			cm.Add()
			go func(wsConfigPath string) {
				var wsPath = path.Dir(wsConfigPath)

				// TODO: error handling
				var wsCfg, err = workspace.LoadConfig(wsConfigPath)

				if err == nil {
					workspacesConfigsMap.Store(wsPath, wsCfg)
				}
				cm.Done()
			}(wsConfigPath)
		}
	}

	cm.Wait()

	return workspacesConfigsMap
}
