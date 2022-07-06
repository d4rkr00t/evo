package runner

import (
	"evo/internal/context"
	"evo/internal/goccm"
	"evo/internal/project"
	"evo/internal/stats"
)

func CacheWorkspacesStates(ctx *context.Context, proj *project.Project) {
	defer ctx.Tracer.Event("caching workspaces states").Done()
	ctx.Stats.Start("caching workspaces states", stats.MeasureKindStage)

	if len(proj.WorkspacesNames) == 0 {
		ctx.Stats.Stop("caching workspaces states")
		return
	}

	var ccm = goccm.New(ctx.Concurrency)

	for _, wsName := range proj.WorkspacesNames {
		ccm.Wait()
		go func(wsName string) {
			defer ccm.Done()
			var ws, _ = proj.Load(wsName)
			ws.CacheState(&ctx.Cache)
		}(wsName)
	}

	ccm.WaitAllDone()

	ctx.Stats.Stop("caching workspaces states")
}
