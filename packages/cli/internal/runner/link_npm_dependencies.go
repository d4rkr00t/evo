package runner

import (
	"evo/internal/ccm"
	"evo/internal/context"
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

	if len(proj.WorkspacesNames) == 0 {
		lg.Debug().EndEmpty(ctx.Stats.Stop("link npm dependencies"))
		return nil
	}

	var ccm = ccm.New(ctx.Concurrency)

	for _, wsName := range proj.WorkspacesNames {
		ccm.Add()
		go func(ws_name string) {
			defer ctx.Tracer.Event(fmt.Sprintf("linkin npm dependencies for %s", ws_name)).Done()
			defer ccm.Done()
			npm.LinkNpmDependencies(proj, ws_name)
		}(wsName)
	}

	ccm.Wait()

	lg.Debug().EndEmpty(ctx.Stats.Stop("link npm dependencies"))

	return nil
}
