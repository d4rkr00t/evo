package lib

import "fmt"

func GetTopLevelRulesNames(root_config *Config) []string {
	return root_config.GetRulesNames()
}

func GetScoppedRulesNames(ctx *Context, scope string) ([]string, error) {
	var targets = []string{}
	var wm, _ = NewWorkspaceMap(ctx.root, &ctx.config, &ctx.cache)
	var ws, ok = wm.Load(scope)
	if !ok {
		return targets, fmt.Errorf("no workspace named %s", scope)
	}
	for name := range ctx.config.GetAllRulesForWS(ctx.root, ws.Path) {
		targets = append(targets, name)
	}
	return targets, nil
}
