package lib

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func Run(ctx Context) {
	ctx.stats.StartMeasure("total", MEASURE_KIND_STAGE)
	defer print_total_time(&ctx)

	os.Setenv("PATH", GetNodeModulesBinPath(ctx.root)+":"+os.ExpandEnv("$PATH"))
	os.Setenv("ROOT", ctx.root)

	ctx.logger.Log()
	ctx.logger.LogWithBadge("cwd", "   "+ctx.cwd)
	ctx.logger.LogWithBadge("target", color.CyanString(ctx.target))

	should_continue, _ := install_dependencies_step(&ctx)
	if !should_continue {
		return
	}

	should_continue, workspaces, _, updated_ws, affected_ws := invalidate_workspaces_step(&ctx)
	if !should_continue {
		return
	}

	should_continue, _ = linking_step(&ctx, &workspaces, &updated_ws)
	if !should_continue {
		return
	}

	run_step(&ctx, &workspaces, &updated_ws, &affected_ws)
}

func print_total_time(ctx *Context) {
	ctx.logger.Log()
	var task_parallel_time = ctx.stats.GetMeasure("runtasks").duration
	var task_seq_time = ctx.stats.GetTasksSumDuration()
	var diff = task_seq_time - task_parallel_time
	ctx.logger.LogWithBadgeVerbose(
		"Tasks time",
		color.HiBlackString(
			"%s %s | %s %s |",
			"seq time:",
			task_seq_time.String(),
			"concurent time:",
			task_parallel_time.String(),
		),
		color.GreenString(diff.String()+" saved"),
	)
	ctx.logger.LogWithBadge("Total time", color.GreenString(ctx.stats.StopMeasure("total").String()))
}

func install_dependencies_step(ctx *Context) (bool, error) {
	if ctx.root_pkg_json.Invalidate(&ctx.cache) || !IsNodeModulesExist(ctx.root) {
		ctx.stats.StartMeasure("install", MEASURE_KIND_STAGE)
		var install_lg = ctx.logger.CreateGroup()
		install_lg.Start("Installing dependencies...")

		var err = InstallNodeDeps(ctx.root, &install_lg)

		if err != nil {
			install_lg.ErrorWithBadge("pnpm", err.Error())
			install_lg.End(ctx.stats.StopMeasure("install"))
			return false, err
		}

		ctx.root_pkg_json.CacheState(&ctx.cache)
		install_lg.End(ctx.stats.StopMeasure("install"))
	}

	return true, nil
}

func invalidate_workspaces_step(ctx *Context) (bool, WorkspacesMap, DepGraph, map[string]string, map[string]string) {
	ctx.stats.StartMeasure("invalidate", MEASURE_KIND_STAGE)
	var invalidate_lg = ctx.logger.CreateGroup()
	invalidate_lg.Start("Invalidating workspaces...")

	var workspaces, ws_err = GetWorkspaces(ctx.root, &ctx.config, &ctx.cache)
	var dep_graph = NewDepGraph(&workspaces)

	if ws_err != nil {
		invalidate_lg.ErrorWithBadge("error", ws_err.Error())
		invalidate_lg.End(ctx.stats.StopMeasure("invalidate"))
		return false, workspaces, dep_graph, nil, nil
	}

	if err := ValidateExternalDeps(&workspaces, ctx.root_pkg_json); err != nil {
		invalidate_lg.ErrorWithBadge("error", err.Error())
		invalidate_lg.End(ctx.stats.StopMeasure("invalidate"))
		return false, workspaces, dep_graph, nil, nil
	}

	if err := ValidateDepsGraph(&dep_graph); err != nil {
		invalidate_lg.ErrorWithBadge("error", err.Error())
		invalidate_lg.End(ctx.stats.StopMeasure("invalidate"))
		return false, workspaces, dep_graph, nil, nil
	}

	var updated_ws = InvalidateWorkspaces(&workspaces, ctx.target, &ctx.cache)

	if len(updated_ws) > 0 {
		invalidate_lg.LogWithBadge(
			"updated",
			color.CyanString(fmt.Sprint((len(updated_ws)))),
			"of",
			color.CyanString(fmt.Sprint((len(workspaces)))),
			"workspaces",
		)

		invalidate_lg.LogVerbose("")
		invalidate_lg.LogVerbose("Calculating affected workspaces...")
		var affected_ws = dep_graph.GetAffected(&workspaces, &updated_ws)
		invalidate_lg.LogWithBadge(
			"affected",
			color.CyanString(fmt.Sprint((len(affected_ws)))),
			"of",
			color.CyanString(fmt.Sprint((len(workspaces)))),
			"workspaces",
		)
		invalidate_lg.End(ctx.stats.StopMeasure("invalidate"))

		return true, workspaces, dep_graph, updated_ws, affected_ws
	}

	invalidate_lg.Log("Everything is up-to-date.")
	invalidate_lg.End(ctx.stats.StopMeasure("invalidate"))
	return false, workspaces, dep_graph, updated_ws, map[string]string{}
}

func linking_step(ctx *Context, workspaces *WorkspacesMap, updated_ws *map[string]string) (bool, error) {
	ctx.stats.StartMeasure("linking", MEASURE_KIND_STAGE)
	var linking_lg = ctx.logger.CreateGroup()
	linking_lg.Start("Linking workspaces...")
	LinkWorkspaces(ctx.root, workspaces, updated_ws)
	linking_lg.EndCollapsed(ctx.stats.StopMeasure("linking"))
	return true, nil
}

func run_step(ctx *Context, workspaces *WorkspacesMap, updated_ws *map[string]string, affected_ws *map[string]string) (bool, error) {
	ctx.stats.StartMeasure("run", MEASURE_KIND_STAGE)
	var run_lg = ctx.logger.CreateGroup()

	run_lg.Start(fmt.Sprintf("Running target → %s", color.CyanString(ctx.target)))

	var tasks = CreateTasksFromWorkspaces(
		ctx.target,
		workspaces,
		updated_ws,
		affected_ws,
		&ctx.config,
		&run_lg,
	)

	if len(tasks) > 0 {
		run_lg.LogVerbose("Executing tasks...")
		RunTasks(ctx, &tasks, &run_lg)
	} else {
		run_lg.Warn("No tasks found, skipping...")
	}

	run_lg.End(ctx.stats.StopMeasure("run"))
	return false, nil
}
