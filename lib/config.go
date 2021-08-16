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

func (c Config) GetAllRulesForWS(ws_path string) map[string]Rule {
	var rules = map[string]Rule{}

	// Adding default rules
	for name, rule := range c.Rules {
		rules[name] = rule
	}

	// Adding rule overrides
	for group_name, group := range c.Overrides {
		if val, _ := doublestar.Match(group_name, ws_path); val {
			for name, rule := range group.Rules {
				rules[name] = rule
			}
		}
	}

	return rules
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
