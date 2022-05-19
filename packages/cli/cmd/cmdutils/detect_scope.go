package cmdutils

import (
	"evo/internal/workspace"
	"path"
)

func DetectScopeFromCWD(rootPath string, cwd string) []string {
	if rootPath != cwd {
		var wsConfig, err = workspace.LoadConfig(path.Join(cwd, workspace.WorkspaceConfigFileName))

		if err == nil {
			return []string{wsConfig.Name}
		}
	}

	return []string{}
}
