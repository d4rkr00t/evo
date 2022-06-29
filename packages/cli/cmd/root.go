package cmd

import (
	"fmt"
	"os"
	"runtime"

	"evo/cmd/version"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "evo",
	Short:             "Build orchestration tool.",
	Long:              `A fresh take on monorepo tooling.`,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	Run: func(cmd *cobra.Command, args []string) {
		var version_flag, _ = cmd.Flags().GetBool("version")
		if version_flag {
			fmt.Println(version.Version)
		} else {
			root := cmd.Root()
			root.SetArgs([]string{"--help"})
			root.Execute()
		}
	},
}

func Execute() {
	rootCmd.PersistentFlags().BoolP("version", "", false, "Version")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolP("debug", "", false, "Debug output")

	RunCmd.PersistentFlags().BoolP("ci", "", false, "Indicates that the command is running in CI")
	RunCmd.PersistentFlags().StringSlice("scope", []string{}, "Scope run to specified packages")
	RunCmd.PersistentFlags().Int("concurrency", runtime.NumCPU()-1, "Number of concurrently running tasks, defaults to a number of CPUs")
	RunCmd.PersistentFlags().String("cwd", "", "Override CWD")
	RunCmd.PersistentFlags().String("since", "", "Use git diff to determine what workspaces have changed since a merge-base")
	RunCmd.PersistentFlags().String("tracing", "", "Output tracing data")
	RunCmd.PersistentFlags().Lookup("tracing").NoOptDefVal = "evo-tracing-output.trace"

	ShowHashCmd.PersistentFlags().String("cwd", "", "Override CWD")

	rootCmd.AddCommand(RunCmd)
	rootCmd.AddCommand(ShowHashCmd)
	rootCmd.AddCommand(ShowTargetsCmd)
	rootCmd.AddCommand(ShowScopesCmd)
	rootCmd.AddCommand(ShowAffectedCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
