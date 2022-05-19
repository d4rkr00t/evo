package show

import (
	"evo/internal/context"
	"evo/internal/project"
	"evo/internal/runner"
	"evo/internal/stats"
)

func Affected(ctx *context.Context, targetName string) error {
	ctx.Stats.Start("show-affected", stats.MeasureKindStage)
	ctx.Logger.Log()
	ctx.Logger.Badge("root").Log(" " + ctx.Root)
	ctx.Logger.Badge("query").Log("show affected workspaces")

	var proj, err = project.NewProject(ctx.ProjectConfigPath)
	if err != nil {
		return err
	}

	err = runner.AugmentDependencies(ctx, &proj)
	if err != nil {
		return err
	}

	runner.BuildDependencyGraph(ctx, &proj)

	err = runner.ValidateDependencyGraph(ctx, &proj)
	if err != nil {
		return err
	}

	runner.InvalidateProjects(ctx, &proj)

	var lg = ctx.Logger.CreateGroup()
	var affected = proj.GetAffectedWorkspaces(ctx, []string{targetName})

	if len(affected) == 0 {
		lg.Start("No affected workspaces")
		lg.EndPlainEmpty()
		return nil
	}

	lg.Start("Affected workspaces:")

	for _, wsName := range affected {
		lg.Log("–", wsName)
	}

	lg.End(ctx.Stats.Stop("show-affected"))

	return nil
}
