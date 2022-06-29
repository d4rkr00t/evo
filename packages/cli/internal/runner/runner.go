package runner

import (
	"evo/internal/context"
	"evo/internal/integrations/npm"
	"evo/internal/reporter"
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

	// TODO: Get rid of this
	os.Setenv("PATH", npm.GetNodeModulesBinPath(ctx.Root)+":"+os.ExpandEnv("$PATH"))
	os.Setenv("ROOT", ctx.Root)

	ctx.Logger.Badge("root").Log(ctx.Root)
	if len(ctx.Scope) > 0 {
		ctx.Logger.Badge("scope").Log(color.YellowString(strings.Join(ctx.Scope, ", ")))
	}
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
		err = ValidateScopes(&proj, &scope)
		if err != nil {
			return err
		}

		proj.ReduceToScope(scope)
		if ctx.Reporter.Output == reporter.ReporterOutputOnlyErrors {
			ctx.Reporter.SetOutput(reporter.ReporterOutputStreamTopLevel)
		}
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
	ctx.Stats.Stop("total")

	if err != nil {
		ctx.Reporter.FailRun(ctx.Stats.Get("total").Duration, &taskGraph)
		return err
	}

	ctx.Reporter.SuccessRun(&ctx.Stats, &taskGraph)
	return nil
}
