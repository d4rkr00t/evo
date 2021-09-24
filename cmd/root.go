package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "evo",
	Short:             "Build orchestration tool.",
	Long:              `A fresh take on monorepo tooling.`,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
}

func Execute() {
	RunCmd.PersistentFlags().StringSlice("scope", []string{}, "Scope run to specified target packages")
	RunCmd.PersistentFlags().String("cwd", "", "Override CWD")
	ShowHashCmd.PersistentFlags().String("cwd", "", "Override CWD")
	ShowRulesCmd.PersistentFlags().String("cwd", "", "Override CWD")

	rootCmd.AddCommand(RunCmd)
	rootCmd.AddCommand(ShowHashCmd)
	rootCmd.AddCommand(ShowRulesCmd)
	rootCmd.AddCommand(ShowAffectedCmd)
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
