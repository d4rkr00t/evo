package cmd

import (
	"errors"
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

var ShowScopesCmd = &cobra.Command{
	Use:   "show-scope <workspace name>",
	Short: "Show the scope for a workspace",
	Long:  "Show the scope for a workspace which is a list of all reachable dependency packages",
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

		var err = show.Scope(&ctx, pkgName)

		if err != nil {
			logger.Log(err.Error())
			os.Exit(1)
		}
	},
}