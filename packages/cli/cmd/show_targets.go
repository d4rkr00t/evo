package cmd

import (
	"errors"
	"evo/cmd/cmdutils"
	"evo/internal/cache"
	"evo/internal/context"
	"evo/internal/logger"
	"evo/internal/project"
	"evo/internal/show"
	"evo/internal/stats"
	"evo/internal/tracer"
	"os"
	"path"
	"runtime"

	"github.com/spf13/cobra"
)

var ShowTargetsCmd = &cobra.Command{
	Use:   "show-targets <workspace name>",
	Short: "Show all targets for a workspace with overrides",
	Long:  "Show all targets for a workspace with overrides",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("workspace name is required")
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
		var pkgName = args[0]

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

		var ctx = context.Context{
			Root:              rootPath,
			Cwd:               cwd,
			ProjectConfigPath: projectConfigPath,
			Targets:           []string{},
			Concurrency:       runtime.NumCPU() - 1,
			Logger:            logger,
			Stats:             stats.New(),
			Tracer:            tracer,
			Cache:             cache,
			Scope:             []string{},
		}

		var err = show.Targets(&ctx, pkgName)

		if err != nil {
			logger.Log(err.Error())
			os.Exit(1)
		}
	},
}
