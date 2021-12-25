package lib

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func ShowHash(ctx Context, ws_name string) error {
	ctx.stats.StartMeasure("show-hash", MEASURE_KIND_STAGE)
	ctx.logger.Log()
	ctx.logger.LogWithBadge("cwd", "   "+ctx.cwd)
	ctx.logger.LogWithBadge("query", " show hash of", ws_name)

	var wm, err = NewWorkspaceMap(ctx.root, &ctx.config, &ctx.cache)

	if err != nil {
		return err
	}

	var ws, ok = wm.Load(ws_name)

	if !ok {
		ctx.logger.Log("  Package", ws_name, "not found!")
		return errors.New(fmt.Sprint("  Package", ws_name, "not found!"))
	}

	wm.RehashAll(&ctx)
	ws.Rehash(&wm)

	var lg = ctx.logger.CreateGroup()
	lg.Start("Package hash consists of:")

	lg.Log("Files:", color.HiBlackString(ws.FilesHash))
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
	lg.Log("Rules:", color.HiBlackString(ws.RulesHash))
	var rules = get_rules_names(&ws.Rules)
	for _, rule := range rules {
		lg.Log("–", rule)
	}

	lg.Log()
	lg.Log("Hash:")
	lg.Log("–", ws.hash)

	lg.End(ctx.stats.StopMeasure("show-hash"))

	return nil
}

func ShowRules(ctx Context, ws_name string) error {
	ctx.stats.StartMeasure("show-rules", MEASURE_KIND_STAGE)
	ctx.logger.Log()
	ctx.logger.LogWithBadge("cwd", "   "+ctx.cwd)
	ctx.logger.LogWithBadge("query", " show rules for", ws_name)

	var wm, err = NewWorkspaceMap(ctx.root, &ctx.config, &ctx.cache)

	if err != nil {
		return err
	}

	var ws, ok = wm.Load(ws_name)
	if !ok {
		ctx.logger.Log("  Package", ws_name, "not found!")
		return errors.New(fmt.Sprint("  Package", ws_name, "not found!"))
	}

	ctx.logger.Log()
	ctx.logger.Log("  All rules for a package:")

	for rule_name, rule := range ws.Rules {
		var lg = ctx.logger.CreateGroup()

		lg.Start(rule_name)

		lg.Badge("cwd").Info("  ", rule.Cmd)
		lg.Badge("cache").Info("", fmt.Sprint(rule.CacheOutput))
		if len(rule.Deps) > 0 {
			lg.Badge("deps").Info(" ", strings.Join(rule.Deps, " | "))
		}

		lg.EndPlain()
	}

	return nil
}

func ShowAffected(ctx Context, target []string) error {
	ctx.stats.StartMeasure("show-affected", MEASURE_KIND_STAGE)
	ctx.logger.Log()
	ctx.logger.LogWithBadge("cwd", "   "+ctx.cwd)
	ctx.logger.LogWithBadge("query", " show affected")

	var wm, err = NewWorkspaceMap(ctx.root, &ctx.config, &ctx.cache)

	if err != nil {
		return err
	}

	wm.RehashAll(&ctx)
	wm.Invalidate(&ctx)

	var lg = ctx.logger.CreateGroup()
	lg.Start("Affected packages:")

	wm.updated.Each(func(key interface{}) bool {
		var ws, _ = wm.Load(key.(string))
		lg.Badge(ws.Name).Info(ws.hash)
		return false
	})

	lg.End(ctx.stats.StopMeasure("show-affected"))

	return nil
}

func ShowScope(ctx Context, target string) error {
	ctx.stats.StartMeasure("show-scope", MEASURE_KIND_STAGE)
	ctx.logger.Log()
	ctx.logger.LogWithBadge("cwd", "   "+ctx.cwd)
	ctx.logger.LogWithBadge("query", " show scope for", target)

	var wm, err = NewWorkspaceMap(ctx.root, &ctx.config, &ctx.cache)

	if err != nil {
		return err
	}

	wm.ReduceToScope([]string{target})

	var lg = ctx.logger.CreateGroup()
	lg.Start("Packages in scope:")

	wm.workspaces.Range(func(key, value interface{}) bool {
		var ws_name = key.(string)
		lg.Log("–", ws_name)
		return true
	})

	lg.End(ctx.stats.StopMeasure("show-scope"))

	return nil
}