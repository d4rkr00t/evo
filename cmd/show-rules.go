package cmd

import (
	"errors"
	"evo/main/lib"
	"evo/main/lib/cache"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var ShowRulesCmd = &cobra.Command{
	Use:   "show-rules <package name>",
	Short: "Show all rules for a package with overrides",
	Long:  "Show all rules for a package with overrides",
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
		var pkg_json = lib.NewPackageJson(path.Join(cwd, "package.json"))
		var ctx = lib.NewContext(
			cwd,
			cwd,
			"",
			pkg_json,
			cache.NewCache(cwd),
			lib.NewLogger(verbose),
			lib.NewStats(),
			pkg_json.GetConfig(),
		)

		var err = lib.ShowRules(ctx, args[0])

		if err != nil {
			os.Exit(1)
		}
	},
}
