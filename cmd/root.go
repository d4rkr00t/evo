package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "scu",
	Short:             "Build orchestration tool.",
	Long:              `A fresh take on monorepo tooling.`,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
}

func Execute() {
	BuildCmd.PersistentFlags().String("cwd", "", "Override CWD")

	rootCmd.AddCommand(BuildCmd)
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
