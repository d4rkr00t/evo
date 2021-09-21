package cmd

import (
	"errors"
	"evo/main/lib"
	"evo/main/lib/cache"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var ShowHashCmd = &cobra.Command{
	Use:   "show-hash <package name>",
	Short: "Show what's included in a hash for a package",
	Long:  "Show what's included in a hash for a package",
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
		var logger = lib.NewLogger(verbose)

		if err == nil {
			var ctx = lib.NewContext(
				path.Dir(root_pkg_json.Path),
				cwd,
				args,
				root_pkg_json,
				cache.NewCache(cwd),
				logger,
				lib.NewStats(),
				root_pkg_json.GetConfig(),
			)

			err = lib.ShowHash(ctx, args[0])
		} else {
			logger.Log("Error: Not in evo project!")
		}

		if err != nil {
			os.Exit(1)
		}
	},
}
