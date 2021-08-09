package lib

import "github.com/bmatcuk/doublestar/v4"

type Config struct {
	Workspaces []string
	Rules      map[string]map[string]Rule
}

type Rule struct {
	Cmd  string
	Deps []string
}

func (c Config) GetRule(name string, ws_path string) Rule {
	for r_name, rule := range c.Rules {
		if r_name == "default" {
			continue
		}

		if val, _ := doublestar.Match(r_name, ws_path); val {
			return rule[name]
		}
	}

	return c.Rules["default"][name]
}
