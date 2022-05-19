package project

import (
	"encoding/json"
	"evo/internal/errors"
	"evo/internal/fsutils"
	"evo/internal/target"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

const ProjectConfigFileName = ".evo.json"

type ProjectConfig struct {
	Workspaces []string
	Commands   map[string]string
	Targets    target.TargetsMap
	Excludes   []string
	Overrides  map[string]ProjectConfigOverride
}

type ProjectConfigOverride struct {
	Targets  target.TargetsMap
	Excludes []string
}

func LoadConfig(configPath string) (ProjectConfig, error) {
	var cfg ProjectConfig
	var dat, err = ioutil.ReadFile(configPath)
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(dat, &cfg)
	return cfg, err
}

func FindProjectConfig(cwd string) (string, error) {
	for {
		var maybecfgPath = path.Join(cwd, ProjectConfigFileName)

		if fsutils.Exist(maybecfgPath) {
			return maybecfgPath, nil
		}

		if cwd == path.Dir(cwd) {
			break
		}

		cwd = path.Dir(cwd)
	}

	return "", errors.New(errors.ErrorProjectConfigNotFound, fmt.Sprintf("Evo '%s' config not found. Not an evo project.", ProjectConfigFileName))
}

func (c *ProjectConfig) GetExcludes(rootPath string, wsPath string) []string {
	var excludes = append([]string{}, c.Excludes...)

	for groupName, group := range c.Overrides {
		var absGroupPath = path.Join(rootPath, groupName)
		if strings.HasPrefix(wsPath, absGroupPath) {
			if len(group.Excludes) > 0 {
				excludes = append(excludes, group.Excludes...)
			}
		}
	}

	return excludes
}
