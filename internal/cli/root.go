package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagCwd     string
	flagConfig  string
	flagTimeout string
)

// RootCmd is the top-level command.
var RootCmd = &cobra.Command{
	Use:   "onboardci",
	Short: "OnboardCI verifies open-source repository setup and testing instructions.",
	Long: `OnboardCI parses setup and test instructions from your documentation (README.md,
CONTRIBUTING.md) or .onboardci.yml, and executes them in an isolated Docker container
to guarantee first-time contributors can successfully onboard.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	RootCmd.PersistentFlags().StringVar(&flagCwd, "cwd", ".", "Target repository directory")
	RootCmd.PersistentFlags().StringVar(&flagConfig, "config", "", "Path to .onboardci.yml config file")
	RootCmd.PersistentFlags().StringVar(&flagTimeout, "timeout", "10m", "Execution timeout (e.g. 10m, 300s)")

	RootCmd.AddCommand(checkCmd)
	RootCmd.AddCommand(versionCmd)
}

// Execute runs the root CLI command.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
