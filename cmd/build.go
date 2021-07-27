package cmd

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"

	"scu/main/runner"
)

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a project",
	Long:  "Build a project",
	Run: func(cmd *cobra.Command, args []string) {
		var cwd, _ = os.Getwd()
		var r = runner.NewRunner(cwd)
		fmt.Println("Build: " + r.GetCwd())
		spew.Dump(r)
	},
}
