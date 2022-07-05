package project

import (
	"evo/internal/goccm"
	"evo/internal/workspace"
	"path"
	"path/filepath"
	"runtime"
	"sync"
)

func DiscoverWorkspaces(rootPath string, workspacesGlobsList []string) sync.Map {
	var workspacesConfigsMap sync.Map
	// TODO: use concurency settings
	var ccm = goccm.New(runtime.NumCPU())
	ccm.Wait()

	for _, wc := range workspacesGlobsList {
		var wsGlob = path.Join(rootPath, wc, workspace.WorkspaceConfigFileName)
		var matches, _ = filepath.Glob(wsGlob)

		for _, wsConfigPath := range matches {
			ccm.Wait()
			go func(wsConfigPath string) {
				defer ccm.Done()
				var wsPath = path.Dir(wsConfigPath)

				// TODO: error handling
				var wsCfg, err = workspace.LoadConfig(wsConfigPath)

				if err == nil {
					workspacesConfigsMap.Store(wsPath, wsCfg)
				}
			}(wsConfigPath)
		}
	}

	ccm.Done()
	ccm.WaitAllDone()

	return workspacesConfigsMap
}
