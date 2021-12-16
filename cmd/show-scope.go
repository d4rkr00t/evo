package cmd

import (
	"errors"
	"evo/main/lib"
	"evo/main/lib/cache"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var ShowScopeCmd = &cobra.Command{
	Use:   "show-scope <package name>",
	Short: "Show the scope for a package",
	Long:  "Show the scope for a package which is list of all reachable dependency packages",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("package name is required")
		}
		return nil

	},
	Run: func(cmd *cobra.Command, args []string) {
		var cwd, cwd_err = cmd.Flags().GetString("cwd")
		var os_cwd, _ = os.Getwd()
		if cwd_err != nil {
			cwd = os_cwd
		} else {
			cwd = path.Join(os_cwd, cwd)
		}

		var verbose, _ = cmd.Flags().GetBool("verbose")
		var root_pkg_json, err = lib.FindRootPackageJson(cwd)
		var root_path = path.Dir(root_pkg_json.Path)
		var logger = lib.NewLogger(verbose)

		if err == nil {
			var ctx = lib.NewContext(
				root_path,
				cwd,
				[]string{},
				args,
				1,
				root_pkg_json,
				cache.NewCache(root_path),
				logger,
				lib.NewStats(),
				root_pkg_json.GetConfig(),
			)

			err = lib.ShowScope(ctx, args[0])
		} else {
			logger.Log("Error: Not in evo project!")
		}

		if err != nil {
			os.Exit(1)
		}
	},
}
