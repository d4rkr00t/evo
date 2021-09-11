package lib

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

func Run(ctx Context) error {
	ctx.stats.StartMeasure("total", MEASURE_KIND_STAGE)
	defer print_total_time(&ctx)

	os.Setenv("PATH", GetNodeModulesBinPath(ctx.root)+":"+os.ExpandEnv("$PATH"))
	os.Setenv("ROOT", ctx.root)

	ctx.logger.LogWithBadge("root", "   "+ctx.cwd)
	ctx.logger.LogWithBadge("targets", color.CyanString(strings.Join(ctx.target, ", ")))

	should_continue, err := install_dependencies_step(&ctx)
	if !should_continue {
		return err
	}

	should_continue, wm, err := invalidate_workspaces_step(&ctx)
	if !should_continue {
		return err
	}

	should_continue, err = linking_step(&ctx, &wm)
	if !should_continue {
		return err
	}

	_, err = run_step(&ctx, &wm)

	return err
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
			install_lg.Badge("pnpm").Error(err.Error())
			install_lg.End(ctx.stats.StopMeasure("install"))
			return false, err
		}

		ctx.root_pkg_json.CacheState(&ctx.cache)
		install_lg.End(ctx.stats.StopMeasure("install"))
	}

	return true, nil
}

func invalidate_workspaces_step(ctx *Context) (bool, WorkspacesMap, error) {
	ctx.stats.StartMeasure("invalidate", MEASURE_KIND_STAGE)
	var invalidate_lg = ctx.logger.CreateGroup()
	invalidate_lg.Start("Invalidating workspaces...")

	var wm, ws_err = NewWorkspaceMap(ctx.root, &ctx.config, &ctx.cache)

	if ws_err != nil {
		invalidate_lg.Badge("error").Error(ws_err.Error())
		invalidate_lg.End(ctx.stats.StopMeasure("invalidate"))
		return false, wm, ws_err
	}

	if err := ValidateExternalDeps(&wm, ctx.root_pkg_json); err != nil {
		invalidate_lg.Badge("error").Error(err.Error())
		invalidate_lg.End(ctx.stats.StopMeasure("invalidate"))
		return false, wm, err
	}

	if err := ValidateDepsGraph(&wm.dep_graph); err != nil {
		invalidate_lg.Badge("error").Error(err.Error())
		invalidate_lg.End(ctx.stats.StopMeasure("invalidate"))
		return false, wm, err
	}

	wm.Invalidate(ctx.target)

	if len(wm.updated) > 0 {
		invalidate_lg.Badge("updated").Info(
			color.CyanString(fmt.Sprint((len(wm.updated)))),
			"of",
			color.CyanString(fmt.Sprint((len(wm.workspaces)))),
			"workspaces",
		)
		wm.GetAffected()
		invalidate_lg.Badge("affected").Info(
			color.CyanString(fmt.Sprint((len(wm.affected)))),
			"of",
			color.CyanString(fmt.Sprint((len(wm.workspaces)))),
			"workspaces",
		)
		invalidate_lg.End(ctx.stats.StopMeasure("invalidate"))

		return true, wm, nil
	}

	invalidate_lg.Log("Everything is up-to-date.")
	invalidate_lg.End(ctx.stats.StopMeasure("invalidate"))
	return false, wm, nil
}

func linking_step(ctx *Context, wm *WorkspacesMap) (bool, error) {
	ctx.stats.StartMeasure("linking", MEASURE_KIND_STAGE)
	var linking_lg = ctx.logger.CreateGroup()
	linking_lg.Start("Linking workspaces...")
	LinkWorkspaces(ctx.root, wm)
	linking_lg.End(ctx.stats.StopMeasure("linking"))
	return true, nil
}

func run_step(ctx *Context, workspaces *WorkspacesMap) (bool, error) {
	ctx.stats.StartMeasure("run", MEASURE_KIND_STAGE)
	var run_lg = ctx.logger.CreateGroup()

	run_lg.Start(fmt.Sprintf("Running targets → %s", color.CyanString(strings.Join(ctx.target, ", "))))

	var tasks = CreateTasksFromWorkspaces(
		ctx.target,
		workspaces,
		&ctx.config,
		&run_lg,
	)
	var err error = nil

	if len(tasks) > 0 {
		run_lg.Verbose().Log("Executing tasks...")
		err = RunTasks(ctx, &tasks, workspaces, &run_lg)
	} else {
		run_lg.Warn("No tasks found, skipping...")
	}

	run_lg.End(ctx.stats.StopMeasure("run"))
	return false, err
}
