package runner

import (
	"evo/internal/context"
	"evo/internal/label"
	"evo/internal/reporter"
	"evo/internal/scheduler/basic"
	"evo/internal/stats"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

func Run(ctx *context.Context) error {
	ctx.Stats.Start("total", stats.MeasureKindStage)
	defer ctx.Tracer.Write(&ctx.Logger, ctx.Root)

	os.Setenv("ROOT", ctx.Root)

	ctx.Logger.Badge("root").Log(" ", ctx.Root)
	ctx.Logger.Badge("labels").Log(color.YellowString(label.StringifyLabels(&ctx.Labels)))
	ctx.Logger.Badge("concurrency").Debug().Log(color.YellowString("%d", ctx.Concurrency))

	if len(ctx.ChangedFiles) > 0 {
		ctx.Logger.Badge("changed files:").Debug().Log(fmt.Sprintf("[%d]", len(ctx.ChangedFiles)), strings.Join(ctx.ChangedFiles, ", "))
	}

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

	var scope = label.GetScopeFromLabels(&ctx.Labels)
	if len(ctx.ChangedFiles) > 0 {
		// Reset scope, so only workspaces with changed files are included
		scope = []string{}

		var workspacesWithChangedFiles = proj.GetWorkspacesMatchingFiles(ctx.ChangedFiles)
		scope = append([]string{}, workspacesWithChangedFiles...)
		if len(scope) == 0 {
			ctx.Logger.Log("Nothing changed. Skipping...")
			return nil
		}
	}

	if len(scope) > 0 {
		err = ValidateScopes(&proj, &scope)
		if err != nil {
			return err
		}

		proj.ReduceToScope(scope)
		if ctx.Reporter.Output == reporter.ReporterOutputOnlyErrors {
			ctx.Reporter.SetOutput(reporter.ReporterOutputStreamTopLevel)
		}
	}

	InvalidateProjects(ctx, &proj)

	var taskGraph = CreateTaskGraph(ctx, &proj)
	err = ValidateTaskGraph(ctx, taskGraph)
	if err != nil {
		return err
	}

	LinkNpmDependencies(ctx, &proj)

	err = basic.RunTaskGraph(ctx, &proj, taskGraph)
	ctx.Stats.Stop("total")

	if err != nil {
		ctx.Reporter.FailRun(ctx.Stats.Get("total").Duration, taskGraph)
		return err
	}

	ctx.Reporter.SuccessRun(&ctx.Stats, taskGraph)
	ctx.Cache.CacheData("evo_cache_key", time.Now().String())
	return nil
}
