package show

import (
	"evo/internal/context"
	"evo/internal/errors"
	"evo/internal/label"
	"evo/internal/project"
	"evo/internal/stats"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func Targets(ctx *context.Context, wsName string) error {
	ctx.Stats.Start("show-targets", stats.MeasureKindStage)
	ctx.Logger.Log()
	ctx.Logger.Badge("root").Log(" " + ctx.Root)
	ctx.Logger.Badge("query").Log("show targets for", wsName)

	var proj, err = project.NewProject(ctx.ProjectConfigPath)
	if err != nil {
		return err
	}

	var ws, ok = proj.Load(wsName)
	if !ok {
		return errors.New(errors.ErrorWsNotFound, fmt.Sprint("Workspace", wsName, "not found!"))
	}

	ctx.Logger.Log()

	for targetName, target := range ws.Targets {
		var lg = ctx.Logger.CreateGroup()

		lg.Start(targetName+" |", color.HiYellowString(fmt.Sprintf("evo run %s%s%s", ws.Name, label.Sep, targetName)))

		if len(target.Cmd) > 0 {
			lg.Badge("command").Info(target.Cmd)
		}
		if len(target.Outputs) > 0 {
			lg.Badge("outputs").Info(strings.Join(target.Outputs, " | "))
		}
		if len(target.Deps) > 0 {
			lg.Badge("deps").Info("  ", strings.Join(target.Deps, " | "))
		}

		lg.EndPlain()
	}

	ctx.Stats.Stop("show-targets")
	return nil
}
