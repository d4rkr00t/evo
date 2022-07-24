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
	"evo/internal/label"
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
	Short: "Run a workspaces's target",
	Long:  "Run a workspaces's target",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("target name is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var cwd, cwdErr = cmd.Flags().GetString("cwd")
		var since, _ = cmd.Flags().GetString("since")
		var concurrency, _ = cmd.Flags().GetInt("concurrency")
		var verbose, _ = cmd.Flags().GetBool("verbose")
		var debug, _ = cmd.Flags().GetBool("debug")
		var tracingOutput, _ = cmd.Flags().GetString("tracing")
		var isCI, _ = cmd.Flags().GetBool("ci")
		var maybeTargets = args
		var logger = logger.NewLogger(verbose, debug)
		var tracer = tracer.New()
		var rr = reporter.New(logger)

		var cpuprof, _ = cmd.Flags().GetBool("cpuprof")
		defer cmdutils.CollectProfile(cpuprof)()

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
			os.Exit(1)
		}

		if tracingOutput != "" {
			tracer.SetOut(tracingOutput)
			tracer.Enable()
		}

		var defaultScope = cmdutils.DetectScopeFromCWD(rootPath, cwd)
		var labels, labelsErr = label.GetLablesFromList(maybeTargets, defaultScope)
		if labelsErr != nil {
			fmt.Println(labelsErr.Error())
			os.Exit(1)
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
			Labels:            labels,
			Concurrency:       concurrency,
			ChangedFiles:      changedFiles,
			ChangedOnly:       len(since) > 0,
			Logger:            logger,
			Reporter:          &rr,
			Stats:             stats.New(),
			Tracer:            tracer,
			Cache:             cache,
		}

		var runErr = runner.Run(&ctx)

		if runErr != nil {
			logger.Log()
			logger.Log(runErr.Error())
			os.Exit(1)
		}
	},
}
