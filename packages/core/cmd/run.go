package cmd

import (
	"errors"
	"evo/main/lib"
	"evo/main/lib/cache"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run <command>",
	Short: "Run a project's commmand",
	Long:  "Run a project's command",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("target name is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var cwd, cwd_err = cmd.Flags().GetString("cwd")
		var scope, _ = cmd.Flags().GetStringSlice("scope")
		var concurrency, _ = cmd.Flags().GetInt("concurrency")
		var tracing_output, _ = cmd.Flags().GetString("tracing")

		var os_cwd, _ = os.Getwd()
		if cwd_err != nil {
			cwd = os_cwd
		} else {
			cwd = path.Join(os_cwd, cwd)
		}

		var verbose, _ = cmd.Flags().GetBool("verbose")
		var root_pkg_json, err = lib.FindRootPackageJson(cwd)
		var logger = lib.NewLogger(verbose)
		var root_path = path.Dir(root_pkg_json.Path)
		var tracing = lib.NewTracing()

		if len(scope) == 0 {
			scope = DetectScopeFromCWD(root_path, cwd)
		}

		if tracing_output != "" {
			tracing.SetOut(tracing_output)
			tracing.Enable()
		}

		if err == nil {
			var ctx = lib.NewContext(
				root_path,
				cwd,
				args,
				scope,
				concurrency,
				root_pkg_json,
				cache.NewCache(root_path),
				logger,
				tracing,
				lib.NewStats(),
				root_pkg_json.GetConfig(),
			)

			err = lib.Run(ctx)
		} else {
			logger.Log("Error: Not in evo project!")
		}

		if err != nil {
			os.Exit(1)
		}
	},
}
