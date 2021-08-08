package cmd

import (
	"fmt"
	"os"
	"path"
	"scu/main/lib"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
)

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a project",
	Long:  "Build a project",
	Run: func(cmd *cobra.Command, args []string) {
		var cwd, cwd_err = cmd.Flags().GetString("cwd")
		var os_cwd, _ = os.Getwd()
		if cwd_err != nil {
			cwd = os_cwd
		} else {
			cwd = path.Join(os_cwd, cwd)
		}

		fmt.Println(cwd)

		var r = lib.NewRunner(cwd)
		var verbose, _ = cmd.Flags().GetBool("verbose")
		if verbose {
			spew.Dump(r)
		}
		r.Build()
	},
}
