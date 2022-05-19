package show

import (
	"evo/internal/context"
	"evo/internal/errors"
	"evo/internal/project"
	"evo/internal/runner"
	"evo/internal/stats"
	"evo/internal/target"
	"evo/internal/workspace"
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
)

func Hash(ctx *context.Context, wsName string) error {
	ctx.Stats.Start("show-hash", stats.MeasureKindStage)
	ctx.Logger.Log()
	ctx.Logger.Badge("root").Log(" " + ctx.Root)
	ctx.Logger.Badge("query").Log("show hash of", wsName)

	var proj, err = project.NewProject(ctx.ProjectConfigPath)
	if err != nil {
		return err
	}

	var ws, ok = proj.Load(wsName)
	if !ok {
		ctx.Logger.Log("  Workspace", wsName, "not found!")
		return errors.New(errors.ErrorWsNotFound, fmt.Sprint("  Workspace", wsName, "not found!"))
	}

	err = runner.AugmentDependencies(ctx, &proj)
	if err != nil {
		return err
	}

	proj.ReduceToScope([]string{wsName})
	runner.BuildDependencyGraph(ctx, &proj)

	err = runner.ValidateDependencyGraph(ctx, &proj)
	if err != nil {
		return err
	}

	runner.InvalidateProjects(ctx, &proj)
	ws, _ = proj.Load(wsName)
	var oldWsState, wsCacheErr = ws.RetriveStateFromCache(&ctx.Cache)
	var wsDiff workspace.WorkspacesDiff
	if wsCacheErr == nil {
		wsDiff = workspace.DiffWorkspaces(&oldWsState, ws)
	}

	if wsDiff.Changed {
		var lgChanged = ctx.Logger.CreateGroup()
		lgChanged.Start(color.HiMagentaString("Workspace has changes since last run"))
		if wsDiff.FilesChanged {
			lgChanged.Log("Files:                ", color.YellowString(fmt.Sprintf("%s → %s", oldWsState.FilesHash, ws.FilesHash)))
		}
		if wsDiff.LocalDepsChanged {
			lgChanged.Log("Local dependencies:   ", color.YellowString(fmt.Sprintf("%s → %s", oldWsState.LocalDepsHash, ws.LocalDepsHash)))
		}
		if wsDiff.ExternalDepsChanged {
			lgChanged.Log("External dependencies:", color.YellowString(fmt.Sprintf("%s → %s", oldWsState.ExtDepsHash, ws.ExtDepsHash)))
		}
		if wsDiff.TargetsChanged {
			lgChanged.Log("Targets:              ", color.YellowString(fmt.Sprintf("%s → %s", oldWsState.TargetsHash, ws.TargetsHash)))
		}
		if wsDiff.Changed {
			lgChanged.Log("Workspace:            ", color.YellowString(fmt.Sprintf("%s → %s", oldWsState.Hash, ws.Hash)))
		}
		lgChanged.EndPlain()
	}

	var lg = ctx.Logger.CreateGroup()
	lg.Start("Workspace hash consists of:")
	lg.Log("Files:", color.HiBlackString(ws.FilesHash))
	var files = ws.GetFilesList()
	for _, fileName := range files {
		var filePath, _ = filepath.Rel(ws.Path, fileName)
		lg.Log("–", filePath)
	}

	lg.Log()
	lg.Log("Dependencies:", color.HiBlackString(fmt.Sprintf("[local:%s] [external:%s]", ws.LocalDepsHash, ws.ExtDepsHash)))
	var deps = ws.Deps
	for _, dep := range deps {
		lg.Log(fmt.Sprintf("– %s:%s [%s] [%s]", dep.Name, dep.Version, dep.Type, dep.Provider))
	}

	lg.Log()
	lg.Log("Targets:", color.HiBlackString(ws.TargetsHash))
	var targets = target.GetSortedTargetsNames(&ws.Targets)
	for _, tgt := range targets {
		lg.Log("–", tgt)
	}

	lg.Log()
	lg.Log("Hash:")
	lg.Log("–", ws.Hash)

	lg.End(ctx.Stats.Stop("show-hash"))

	return nil
}
