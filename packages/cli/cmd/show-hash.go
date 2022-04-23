package cmd

import (
	"errors"
	"evo/main/lib"
	"evo/main/lib/cache"
	"fmt"
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
		var root_pkg_json, root_config, err = lib.FindProject(cwd)
		var root_path = path.Dir(root_config.Path)
		var logger = lib.NewLogger(verbose)
		var tracing = lib.NewTracing()

		if err == nil {
			var ctx = lib.NewContext(
				root_path,
				cwd,
				args,
				[]string{},
				[]string{},
				1,
				root_pkg_json,
				cache.NewCache(root_path),
				logger,
				tracing,
				lib.NewStats(),
				root_config,
			)

			err = lib.ShowHash(&ctx, args[0])
		}

		if err != nil {
			logger.Log(fmt.Sprintf("%s", err))
			os.Exit(1)
		}
	},
}
