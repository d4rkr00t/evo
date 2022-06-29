package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"

	"evo/cmd/cmdutils"
	"evo/internal/cache"
	"evo/internal/context"
	"evo/internal/integrations/git"
	"evo/internal/logger"
	"evo/internal/project"
	"evo/internal/reporter"
	"evo/internal/runner"
	"evo/internal/stats"
	"evo/internal/tracer"

	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run <target>",
	Short: "Run a project's target",
	Long:  "Run a project's target",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("target name is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var cwd, cwdErr = cmd.Flags().GetString("cwd")
		var since, _ = cmd.Flags().GetString("since")
		var scope, _ = cmd.Flags().GetStringSlice("scope")
		var concurrency, _ = cmd.Flags().GetInt("concurrency")
		var verbose, _ = cmd.Flags().GetBool("verbose")
		var debug, _ = cmd.Flags().GetBool("debug")
		var tracingOutput, _ = cmd.Flags().GetString("tracing")
		var isCI, _ = cmd.Flags().GetBool("ci")
		var targets = args
		var logger = logger.NewLogger(verbose, debug)
		var tracer = tracer.New()
		var rr = reporter.New(logger)

		if !isCI {
			rr.EnableSpinner()
		}

		if isCI {
			rr.SetOutput(reporter.ReporterOutputCombine)
		} else if verbose || debug {
			rr.SetOutput(reporter.ReporterOutputStreamAll)
		}

		var osCwd, _ = os.Getwd()
		if cwdErr != nil {
			cwd = osCwd
		} else {
			cwd = path.Join(osCwd, cwd)
		}

		var projectConfigPath, projectConfigErr = project.FindProjectConfig(cwd)
		var rootPath = path.Dir(projectConfigPath)

		if projectConfigErr != nil {
			fmt.Println(projectConfigErr.Error())
			logger.Log(fmt.Sprintf("%s", projectConfigErr))
			os.Exit(1)
		}

		if tracingOutput != "" {
			tracer.SetOut(tracingOutput)
			tracer.Enable()
		}

		if len(scope) == 0 {
			scope = cmdutils.DetectScopeFromCWD(rootPath, cwd)
		}

		var cache = cache.New(rootPath, cache.DefaultCacheLocation)
		cache.Setup()

		var changedFiles = []string{}
		if len(since) > 0 {
			changedFiles = git.GetChangedSince(rootPath, since)
		}

		var ctx = context.Context{
			Root:              rootPath,
			Cwd:               cwd,
			ProjectConfigPath: projectConfigPath,
			Targets:           targets,
			Concurrency:       concurrency,
			ChangedFiles:      changedFiles,
			ChangedOnly:       len(since) > 0,
			Logger:            logger,
			Reporter:          &rr,
			Stats:             stats.New(),
			Tracer:            tracer,
			Cache:             cache,
			Scope:             scope,
		}

		var runErr = runner.Run(&ctx)

		if runErr != nil {
			logger.Log()
			logger.Log(runErr.Error())
			os.Exit(1)
		}
	},
}
