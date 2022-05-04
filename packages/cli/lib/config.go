package lib

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type ConfigOverride struct {
	Rules    map[string]Rule
	Excludes []string
}

type Config struct {
	Path       string
	Workspaces []string
	Commands   map[string]string
	Rules      map[string]Rule
	Excludes   []string
	Overrides  map[string]ConfigOverride
}

const CONFIG_FILE_NAME = ".evo.json"

func NewConfig(config_path string) (Config, error) {
	var cfg Config
	var dat, err = ioutil.ReadFile(config_path)
	if err != nil {
		return cfg, err
	}
	json.Unmarshal(dat, &cfg)
	cfg.Path = config_path
	return cfg, nil
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

	var override_groups = []string{}
	for group_name := range c.Overrides {
		override_groups = append(override_groups, group_name)
	}
	sort.Strings(override_groups)

	// Adding rule overrides
	for _, group_name := range override_groups {
		var group = c.Overrides[group_name]
		var abs_group_path = path.Join(root_path, group_name)
		if strings.HasPrefix(ws_path, abs_group_path) {
			for name, rule := range group.Rules {
				rules[name] = expand_cmd(rule)
			}
		}
	}

	return rules
}

func (c Config) GetExcludes(root_path string, ws_path string) []string {
	var excludes = c.Excludes
	var rel_path, _ = filepath.Rel(root_path, ws_path)

	for group_name, group := range c.Overrides {
		if val, _ := doublestar.Match(group_name, rel_path); val {
			if len(group.Excludes) > 0 {
				excludes = group.Excludes
			}
		}
	}

	return excludes
}

func (c Config) GetRulesNames() []string {
	var rules = []string{}
	for name := range c.Rules {
		rules = append(rules, name)
	}
	sort.Strings(rules)
	return rules
}
