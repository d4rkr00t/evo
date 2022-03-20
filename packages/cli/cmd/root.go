package cmd

import (
	"evo/main/lib"
	"fmt"
	"os"
	"runtime"

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
			fmt.Println(lib.Version)
		} else {
			root := cmd.Root()
			root.SetArgs([]string{"--help"})
			root.Execute()
		}
	},
}

func Execute() {
	RunCmd.PersistentFlags().StringSlice("scope", []string{}, "Scope run to specified target packages")
	RunCmd.PersistentFlags().Int("concurrency", runtime.NumCPU(), "Number of concurrently running tasks, defaults to a number of CPUs")
	RunCmd.PersistentFlags().String("cwd", "", "Override CWD")
	RunCmd.PersistentFlags().String("tracing", "", "Output tracing data")
	RunCmd.PersistentFlags().Lookup("tracing").NoOptDefVal = "evo-tracing-output.trace"

	ShowHashCmd.PersistentFlags().String("cwd", "", "Override CWD")

	ShowRulesCmd.PersistentFlags().String("cwd", "", "Override CWD")

	rootCmd.AddCommand(RunCmd)
	rootCmd.AddCommand(ShowHashCmd)
	rootCmd.AddCommand(ShowRulesCmd)
	rootCmd.AddCommand(ShowAffectedCmd)
	rootCmd.AddCommand(ShowScopeCmd)
	rootCmd.PersistentFlags().BoolP("version", "", false, "Version")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
