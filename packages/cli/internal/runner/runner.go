package runner

import (
	"evo/internal/context"
	"evo/internal/integrations/npm"
	"evo/internal/scheduler/basic"
	"evo/internal/stats"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

func Run(ctx *context.Context) error {
	ctx.Stats.Start("total", stats.MeasureKindStage)
	defer ctx.Tracer.Write(&ctx.Logger, ctx.Root)
	defer printTotalTime(ctx)

	// TODO: Get rid of this
	os.Setenv("PATH", npm.GetNodeModulesBinPath(ctx.Root)+":"+os.ExpandEnv("$PATH"))
	os.Setenv("ROOT", ctx.Root)

	ctx.Logger.Badge("root").Log("  ", ctx.Root)
	if len(ctx.Scope) > 0 {
		ctx.Logger.Badge("scope").Log("  " + color.YellowString(strings.Join(ctx.Scope, ", ")))
	}
	ctx.Logger.Badge("targets").Log(color.CyanString(strings.Join(ctx.Targets, ", ")))
	ctx.Logger.Badge("changed files:").Debug().Log(fmt.Sprintf("[%d]", len(ctx.ChangedFiles)), strings.Join(ctx.ChangedFiles, ", "))

	if ctx.ChangedOnly && len(ctx.ChangedFiles) == 0 {
		ctx.Logger.Log("Nothing changed. Skipping...")
		return nil
	}

	var proj, err = CreateProject(ctx)
	if err != nil {
		return err
	}

	err = AugmentDependencies(ctx, &proj)
	if err != nil {
		return err
	}

	var scope = ctx.Scope
	if len(ctx.ChangedFiles) > 0 {
		var workspacesWithChangedFiles = proj.GetWorkspacesMatchingFiles(ctx.ChangedFiles)
		scope = append([]string{}, workspacesWithChangedFiles...)
		if len(scope) == 0 {
			ctx.Logger.Log("Nothing changed. Skipping...")
			return nil
		}
	}

	if len(scope) > 0 {
		proj.ReduceToScope(scope)
	}

	BuildDependencyGraph(ctx, &proj)

	err = ValidateDependencyGraph(ctx, &proj)
	if err != nil {
		return err
	}

	InvalidateProjects(ctx, &proj)
	CacheWorkspacesStates(ctx, &proj)

	var taskGraph = CreateTaskGraph(ctx, &proj)
	err = ValidateTaskGraph(ctx, &taskGraph)
	if err != nil {
		return err
	}

	LinkNpmDependencies(ctx, &proj)

	err = basic.RunTaskGraph(ctx, &proj, &taskGraph)
	if err != nil {
		return err
	}

	return nil
}

func printTotalTime(ctx *context.Context) {
	var taskParallelTime = ctx.Stats.Get("runtasks").Duration
	var taskSeqTime = ctx.Stats.GetTasksSumDuration()
	var diff = taskSeqTime - taskParallelTime

	ctx.Logger.Log()
	ctx.Logger.Badge("Tasks time").Log(
		color.HiBlackString(
			"%s %s | %s %s |",
			"seq time:",
			taskSeqTime.String(),
			"concurent time:",
			taskParallelTime.String(),
		),
		color.GreenString(diff.String()+" saved"),
	)
	ctx.Logger.Badge("Total time").Log(color.GreenString(ctx.Stats.Stop("total").String()))
}
