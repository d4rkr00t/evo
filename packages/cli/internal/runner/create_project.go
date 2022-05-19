package runner

import (
	"evo/internal/context"
	"evo/internal/project"
	"evo/internal/stats"
	"fmt"
)

func CreateProject(ctx *context.Context) (project.Project, error) {
	defer ctx.Tracer.Event("create project").Done()
	ctx.Stats.Start("project create", stats.MeasureKindStage)
	var projLg = ctx.Logger.CreateGroup()
	projLg.Debug().Start("Creating a project...")
	var proj, projErr = project.NewProject(ctx.ProjectConfigPath)
	projLg.Debug().Badge("Number of workspaces").Log(fmt.Sprint(proj.Size()))
	projLg.Debug().End(ctx.Stats.Stop("project create"))

	if projErr != nil {
		return project.Project{}, projErr
	}

	return proj, nil
}
