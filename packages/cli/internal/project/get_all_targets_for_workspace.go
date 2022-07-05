package project

import (
	"evo/internal/target"
	"evo/internal/workspace"
	"path"
	"strings"
)

// TODO: figure out sorting of overrides
func GetAllTargetsForWorkspace(projectPath string, projectConfig *ProjectConfig, wsPath string, wsConfig *workspace.WorkspaceConfig) target.TargetsMap {
	var targets = target.MergeTargets(&target.TargetsMap{}, &projectConfig.Targets)

	var expandCmdFn = func(target target.Target) target.Target {
		var cmd = strings.Split(target.Cmd, " ")[0]
		if expandedCmd, ok := projectConfig.Commands[cmd]; ok {
			target.Cmd = strings.Replace(target.Cmd, cmd, expandedCmd, 1)
		}
		return target
	}

	for groupName := range projectConfig.Overrides {
		var absGroupPath = path.Join(projectPath, groupName)
		if strings.HasSuffix(wsPath, absGroupPath) {
			var group = projectConfig.Overrides[groupName]
			targets = target.MergeTargets(&targets, &group.Targets)
		}
	}

	targets = target.MergeTargets(&targets, &wsConfig.Targets)

	for name, rule := range targets {
		targets[name] = expandCmdFn(rule)
	}

	return targets
}
