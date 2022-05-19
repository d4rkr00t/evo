package workspace

import (
	"encoding/json"
	"evo/internal/target"
	"io/ioutil"
)

const WorkspaceConfigFileName = ".evows.json"

type WorkspaceConfigMap = map[string]WorkspaceConfig

type WorkspaceConfig struct {
	Name     string
	Targets  target.TargetsMap
	Excludes []string
}

func LoadConfig(wsConfigPath string) (WorkspaceConfig, error) {
	var cfg WorkspaceConfig
	var dat, err = ioutil.ReadFile(wsConfigPath)
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(dat, &cfg)
	return cfg, err
}
