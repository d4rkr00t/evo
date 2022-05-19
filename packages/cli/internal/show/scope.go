package show

import (
	"evo/internal/context"
	"evo/internal/project"
	"evo/internal/runner"
	"evo/internal/stats"
)

func Scope(ctx *context.Context, wsName string) error {
	ctx.Stats.Start("show-scope", stats.MeasureKindStage)
	ctx.Logger.Log()
	ctx.Logger.Badge("root").Log(" " + ctx.Root)
	ctx.Logger.Badge("query").Log("show scope for", wsName)

	var proj, err = project.NewProject(ctx.ProjectConfigPath)
	if err != nil {
		return err
	}

	err = runner.AugmentDependencies(ctx, &proj)
	if err != nil {
		return err
	}

	proj.ReduceToScope([]string{wsName})

	var lg = ctx.Logger.CreateGroup()
	lg.Start("Workspaces in scope:")

	for _, wsName := range proj.WorkspacesNames {
		lg.Log("–", wsName)
	}

	lg.End(ctx.Stats.Stop("show-scope"))
	return nil
}
