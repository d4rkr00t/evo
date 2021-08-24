package lib

import (
	"path"
	"strings"

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
	Commands   map[string]string
	Rules      map[string]Rule
	Inputs     ConfigInputs
	Overrides  map[string]ConfigOverride
}

func (c Config) GetAllRulesForWS(root_path string, ws_path string) map[string]Rule {
	var rules = map[string]Rule{}

	var expand_cmd = func(rule Rule) Rule {
		var cmd = strings.Split(rule.Cmd, " ")[0]
		if expanded_cmd, ok := c.Commands[cmd]; ok {
			rule.Cmd = strings.Replace(rule.Cmd, cmd, expanded_cmd, 1)
		}
		return rule
	}

	// Adding default rules
	for name, rule := range c.Rules {
		rules[name] = expand_cmd(rule)
	}

	// Adding rule overrides
	for group_name, group := range c.Overrides {
		var abs_group_path = path.Join(root_path, group_name)
		if strings.HasPrefix(ws_path, abs_group_path) {
			for name, rule := range group.Rules {
				rules[name] = expand_cmd(rule)
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
