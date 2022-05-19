package runner

import (
	"evo/internal/context"
	"evo/internal/goccm"
	"evo/internal/integrations/npm"
	"evo/internal/project"
	"evo/internal/stats"
	"fmt"
)

func LinkNpmDependencies(ctx *context.Context, proj *project.Project) error {
	defer ctx.Tracer.Event("linkin npm dependencies").Done()
	ctx.Stats.Start("link npm dependencies", stats.MeasureKindStage)
	var lg = ctx.Logger.CreateGroup()
	lg.Debug().Start("Linking npm dependencies...")

	var ccm = goccm.New(ctx.Concurrency)
	for _, wsName := range proj.WorkspacesNames {
		ccm.Wait()
		go func(ws_name string) {
			defer ctx.Tracer.Event(fmt.Sprintf("linkin npm dependencies for %s", ws_name)).Done()
			defer ccm.Done()
			npm.LinkNpmDependencies(proj, ws_name)
		}(wsName)
	}
	ccm.WaitAllDone()

	lg.Debug().EndEmpty(ctx.Stats.Stop("link npm dependencies"))

	return nil
}
