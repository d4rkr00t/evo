package cmd

import (
	"evo/main/lib"
	"evo/main/lib/cache"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run <command>",
	Short: "Run a project's commmand",
	Long:  "Run a project's command",
	Run: func(cmd *cobra.Command, args []string) {
		var cwd, cwd_err = cmd.Flags().GetString("cwd")
		var scope, _ = cmd.Flags().GetStringSlice("scope")
		var since, _ = cmd.Flags().GetString("since")
		var concurrency, _ = cmd.Flags().GetInt("concurrency")
		var tracing_output, _ = cmd.Flags().GetString("tracing")

		var os_cwd, _ = os.Getwd()
		if cwd_err != nil {
			cwd = os_cwd
		} else {
			cwd = path.Join(os_cwd, cwd)
		}

		var verbose, _ = cmd.Flags().GetBool("verbose")
		var root_pkg_json, root_config, err = lib.FindProject(cwd)
		var logger = lib.NewLogger(verbose)
		var root_path = path.Dir(root_config.Path)
		var tracing = lib.NewTracing()

		if len(scope) == 0 {
			scope = DetectScopeFromCWD(root_path, cwd)
		}

		if tracing_output != "" {
			tracing.SetOut(tracing_output)
			tracing.Enable()
		}

		if len(args) < 1 {
			logger.Log("Usage: evo run <target>")
			var lg = logger.CreateGroup()
			lg.Start("Available targets:")
			for _, name := range root_config.GetRulesNames() {
				lg.Log(fmt.Sprintf("– %s", name))
			}
			lg.EndPlain()
			return
		}

		var changed_files = []string{}

		if len(since) > 0 {
			changed_files = lib.GetChangedSince(root_path, since)
		}

		if err == nil {
			var ctx = lib.NewContext(
				root_path,
				cwd,
				args,
				scope,
				changed_files,
				concurrency,
				root_pkg_json,
				cache.NewCache(root_path),
				logger,
				tracing,
				lib.NewStats(),
				root_config,
			)

			err = lib.Run(&ctx)
		} else {
			logger.Log("Error: Not in evo project!")
		}

		if err != nil {
			os.Exit(1)
		}
	},
}
