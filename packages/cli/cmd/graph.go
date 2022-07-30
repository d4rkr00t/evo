package cmd

import (
	"evo/cmd/cmdutils"
	"evo/internal/cache"
	"evo/internal/context"
	"evo/internal/graph"
	"evo/internal/label"
	"evo/internal/logger"
	"evo/internal/project"
	"evo/internal/stats"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/spf13/cobra"
)

var GraphCmd = &cobra.Command{
	Use:   "graph <workspace#target?>",
	Short: "Outputs dotviz graph for workspaces",
	Long:  "Outputs dotviz graph for workspaces",
	Run: func(cmd *cobra.Command, args []string) {
		var cwd, _ = os.Getwd()
		var maybeTargets = args
		var logger = logger.NewLogger(false, false)

		var projectConfigPath, projectConfigErr = project.FindProjectConfig(cwd)
		var rootPath = path.Dir(projectConfigPath)

		if projectConfigErr != nil {
			fmt.Println(projectConfigErr.Error())
			logger.Log(fmt.Sprintf("%s", projectConfigErr))
			os.Exit(1)
		}

		var defaultScope = cmdutils.DetectScopeFromCWD(rootPath, cwd)
		var labels, labelsErr = label.GetLablesFromList(maybeTargets, defaultScope)
		if labelsErr != nil {
			fmt.Println(labelsErr.Error())
			os.Exit(1)
		}

		var cache = cache.New(rootPath, cache.DefaultCacheLocation)
		cache.Setup()

		var ctx = context.Context{
			Root:              rootPath,
			Cwd:               cwd,
			ProjectConfigPath: projectConfigPath,
			Labels:            labels,
			Concurrency:       runtime.NumCPU(),
			ChangedFiles:      []string{},
			ChangedOnly:       false,
			Logger:            logger,
			Stats:             stats.New(),
			Cache:             cache,
		}

		var err = graph.Graph(&ctx)

		if err != nil {
			logger.Log()
			logger.Log(err.Error())
			os.Exit(1)
		}
	},
}
