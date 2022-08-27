package cmd

import (
	"errors"
	"evo/cmd/cmdutils"
	"evo/internal/cache"
	"evo/internal/context"
	"evo/internal/label"
	"evo/internal/logger"
	"evo/internal/project"
	"evo/internal/show"
	"evo/internal/stats"
	"evo/internal/tracer"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/spf13/cobra"
)

var ShowAffectedCmd = &cobra.Command{
	Use:   "show-affected <workspace name>::<target>",
	Short: "Show what targets are affected by a change",
	Long:  "Show what targets are affected by a change",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("::target or workspace::target is required")
		}
		return nil

	},
	Run: func(cmd *cobra.Command, args []string) {
		var cwd, cwdErr = cmd.Flags().GetString("cwd")
		var verbose, _ = cmd.Flags().GetBool("verbose")
		var debug, _ = cmd.Flags().GetBool("debug")

		var cpuprof, _ = cmd.Flags().GetBool("cpuprof")
		defer cmdutils.CollectProfile(cpuprof)()

		var logger = logger.NewLogger(verbose, debug)
		var tracer = tracer.New()

		var osCwd, _ = os.Getwd()
		if cwdErr != nil {
			cwd = osCwd
		} else {
			cwd = path.Join(osCwd, cwd)
		}

		var projectConfigPath, projectConfigErr = project.FindProjectConfig(cwd)
		var rootPath = path.Dir(projectConfigPath)

		if projectConfigErr != nil {
			logger.Log(projectConfigErr.Error())
			os.Exit(1)
		}

		var cache = cache.New(rootPath, cache.DefaultCacheLocation)
		cache.Setup()

		var labels, labelsErr = label.GetLablesFromList(args, "*")
		if labelsErr != nil {
			fmt.Println(labelsErr.Error())
			os.Exit(1)
		}

		var ctx = context.Context{
			Root:              rootPath,
			Cwd:               cwd,
			ProjectConfigPath: projectConfigPath,
			Concurrency:       runtime.NumCPU() - 1,
			Logger:            logger,
			Stats:             stats.New(),
			Labels:            []label.Label{labels[0]},
			Tracer:            tracer,
			Cache:             cache,
		}

		var err = show.Affected(&ctx, labels[0])

		if err != nil {
			logger.Log(err.Error())
			os.Exit(1)
		}
	},
}
