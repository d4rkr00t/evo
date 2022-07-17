package runner

import (
	"evo/internal/ccm"
	"evo/internal/context"
	"evo/internal/integrations/npm"
	"evo/internal/project"
	"evo/internal/stats"
	"fmt"
)

func AugmentDependencies(ctx *context.Context, proj *project.Project) error {
	defer ctx.Tracer.Event("discover dependencies").Done()
	ctx.Stats.Start("augment dependencies", stats.MeasureKindStage)
	var lg = ctx.Logger.CreateGroup()
	lg.Debug().Start("Discovering dependencies...")

	if len(proj.WorkspacesNames) == 0 {
		lg.Debug().EndEmpty(ctx.Stats.Stop("augment dependencies"))
		return nil
	}

	var ccm = ccm.New(ctx.Concurrency)
	for _, wsName := range proj.WorkspacesNames {
		ccm.Add()
		go func(wsName string) {
			defer ctx.Tracer.Event(fmt.Sprintf("discover dependencies for %s", wsName)).Done()
			defer ccm.Done()
			// TODO: error handling
			var err = npm.AddNpmDependencies(proj, wsName)
			if err != nil {
				fmt.Println(err.Error())
			}
		}(wsName)
	}

	ccm.Wait()
	lg.Debug().EndEmpty(ctx.Stats.Stop("augment dependencies"))

	return nil
}
