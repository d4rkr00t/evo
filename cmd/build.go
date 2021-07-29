package cmd

import (
	"os"

	"scu/main/lib"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
)

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a project",
	Long:  "Build a project",
	Run: func(cmd *cobra.Command, args []string) {
		var cwd, _ = os.Getwd()
		var r = lib.NewRunner(cwd)
		spew.Dump(r)
		r.Build()
	},
}
