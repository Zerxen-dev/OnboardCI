package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the current release tag.
var Version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of OnboardCI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("OnboardCI v%s (clean-container onboarding verifier)\n", Version)
	},
}
