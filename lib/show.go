package lib

import (
	"fmt"
	"strings"
)

func ShowHash(ctx Context, ws_name string) {
	ctx.stats.StartMeasure("show-hash", MEASURE_KIND_STAGE)
	ctx.logger.Log()
	ctx.logger.LogWithBadge("cwd", "   "+ctx.cwd)
	ctx.logger.LogWithBadge("query", " show hash of", ws_name)

	var wm, _ = NewWorkspaceMap(ctx.root, &ctx.config, &ctx.cache)

	var ws, ok = wm.workspaces[ws_name]
	if !ok {
		ctx.logger.Log("  Package", ws_name, "not found!")
		return
	}

	var lg = ctx.logger.CreateGroup()
	lg.Start("Package hash consists of:")

	lg.Log("Files:")
	var files = ws.get_files()
	for _, file_name := range files {
		lg.Log("–", file_name)
	}

	lg.Log()
	lg.Log("Deps:")
	var deps = ws.Deps
	for dep_name, dep_ver := range deps {
		lg.Log("–", dep_name, ":", dep_ver)
	}

	lg.Log()
	lg.Log("Rules:")
	var rules = ws.get_rules_names()
	for _, rule := range rules {
		lg.Log("–", rule)
	}

	lg.Log()
	lg.Log("Hash:")
	lg.Log("–", ws.Hash(&wm))

	lg.End(ctx.stats.StopMeasure("show-hash"))
}

func ShowRules(ctx Context, ws_name string) {
	ctx.stats.StartMeasure("show-rules", MEASURE_KIND_STAGE)
	ctx.logger.Log()
	ctx.logger.LogWithBadge("cwd", "   "+ctx.cwd)
	ctx.logger.LogWithBadge("query", " show rules for", ws_name)

	var wm, _ = NewWorkspaceMap(ctx.root, &ctx.config, &ctx.cache)

	var ws, ok = wm.workspaces[ws_name]
	if !ok {
		ctx.logger.Log("  Package", ws_name, "not found!")
		return
	}

	ctx.logger.Log()
	ctx.logger.Log("  All rules for a package:")

	for rule_name, rule := range ws.Rules {
		var lg = ctx.logger.CreateGroup()

		lg.Start(rule_name)

		lg.LogWithBadge("cmd", "  ", rule.Cmd)
		lg.LogWithBadge("cache", "", fmt.Sprint(rule.CacheOutput))
		if len(rule.Deps) > 0 {
			lg.LogWithBadge("deps", " ", strings.Join(rule.Deps, " | "))
		}

		lg.EndPlain()
	}

}
