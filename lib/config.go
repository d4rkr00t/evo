package lib

import (
	"github.com/bmatcuk/doublestar/v4"
)

type ConfigOverride struct {
	Rules  map[string]Rule
	Inputs ConfigInputs
}

type ConfigInputs struct {
	Includes []string
	Excludes []string
}

type Config struct {
	Workspaces []string
	Rules      map[string]Rule
	Inputs     ConfigInputs
	Overrides  map[string]ConfigOverride
}

func (c Config) GetRule(name string, ws_path string) Rule {
	for group_name, group := range c.Overrides {
		if val, _ := doublestar.Match(group_name, ws_path); val {
			if val, ok := group.Rules[name]; ok {
				return val
			}
		}
	}

	return c.Rules[name]
}

func (c Config) GetInputs(ws_path string) ([]string, []string) {
	var includes = c.Inputs.Includes
	var excludes = c.Inputs.Excludes

	for group_name, group := range c.Overrides {
		if val, _ := doublestar.Match(group_name, ws_path); val {
			if len(group.Inputs.Includes) > 0 {
				includes = group.Inputs.Includes
			}
			if len(group.Inputs.Excludes) > 0 {
				excludes = group.Inputs.Excludes
			}
		}
	}

	return includes, excludes
}
