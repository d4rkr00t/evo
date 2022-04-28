package cmd

import (
	"evo/main/lib"
	"evo/main/lib/cache"
	"fmt"
	"os"
	"path"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show a list of available targets",
	Long:  "Show a list of available targets for top level or scopped to a package",
	Run: func(cmd *cobra.Command, args []string) {
		var cwd, cwd_err = cmd.Flags().GetString("cwd")
		var concurrency, _ = cmd.Flags().GetInt("concurrency")
		var scope, _ = cmd.Flags().GetStringSlice("scope")

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

		if err == nil {

			if len(scope) == 0 {
				ListTopLevelRules(&root_config, &logger)
			} else {
				var ctx = lib.NewContext(
					root_path,
					cwd,
					args,
					scope,
					[]string{},
					concurrency,
					root_pkg_json,
					cache.NewCache(root_path),
					logger,
					tracing,
					lib.NewStats(),
					root_config,
				)
				err = ListScoppedRules(&ctx, &logger, scope[0])
			}
		}

		if err != nil {
			logger.Log(fmt.Sprintf("%s", err))
			os.Exit(1)
		}
	},
}

func ListTopLevelRules(root_config *lib.Config, logger *lib.Logger) {
	var lg = logger.CreateGroup()
	lg.Start("Available targets:")
	for _, name := range lib.GetTopLevelRulesNames(root_config) {
		lg.Log(fmt.Sprintf("– evo run %s", name))
	}
	lg.EndPlain()
}

func ListScoppedRules(ctx *lib.Context, logger *lib.Logger, scope string) error {
	var lg = logger.CreateGroup()
	lg.Start(fmt.Sprintf("Available targets for %s:", scope))
	var rules, err = lib.GetScoppedRulesNames(ctx, scope)
	if err != nil {
		return err
	}

	for _, name := range rules {
		lg.Log(fmt.Sprintf("– evo run %s %s", name, color.BlackString(fmt.Sprintf("--scope %s", scope))))
	}
	lg.EndPlain()

	return nil
}
