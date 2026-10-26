package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/Zerxen-dev/onboardci/internal/config"
	"github.com/Zerxen-dev/onboardci/internal/extractor"
	"github.com/Zerxen-dev/onboardci/internal/reporter"
	"github.com/Zerxen-dev/onboardci/internal/runner"
)

var (
	flagExecute bool
	flagDryRun  bool
	flagJSON    bool
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Extract and verify onboarding commands from docs or .onboardci.yml",
	Long: `Inspects the target repository for documented setup and test steps.
By default, runs in dry-run mode and lists the extracted commands without executing.
Pass --execute to run commands inside an isolated Docker container.`,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().BoolVar(&flagExecute, "execute", false, "Execute extracted commands inside Docker sandbox")
	checkCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "List extracted commands without executing (default behavior)")
	checkCmd.Flags().BoolVar(&flagJSON, "json", false, "Output machine-readable JSON format")
}

func runCheck(cmd *cobra.Command, args []string) error {
	absCwd, err := filepath.Abs(flagCwd)
	if err != nil {
		return fmt.Errorf("invalid working directory %q: %w", flagCwd, err)
	}

	cfg, configPath, err := config.LoadConfig(absCwd, flagConfig)
	if err != nil {
		return err
	}

	// Override timeout if explicitly provided on CLI
	if cmd.Flags().Changed("timeout") {
		t, err := time.ParseDuration(flagTimeout)
		if err != nil {
			return fmt.Errorf("invalid --timeout flag value %q: %w", flagTimeout, err)
		}
		cfg.Docker.Timeout = t
	}

	var commands []extractor.ExtractedCommand

	// Extraction logic:
	// 1. Prefer .onboardci.yml if present (and defines setup or test commands)
	if configPath != "" && (len(cfg.Setup) > 0 || len(cfg.Test) > 0) {
		for _, s := range cfg.Setup {
			commands = append(commands, extractor.ExtractedCommand{
				SourceFile: configPath,
				Type:       extractor.StepTypeSetup,
				Command:    s,
			})
		}
		for _, t := range cfg.Test {
			commands = append(commands, extractor.ExtractedCommand{
				SourceFile: configPath,
				Type:       extractor.StepTypeTest,
				Command:    t,
			})
		}
	} else {
		// 2. Parse Markdown docs
		extracted, err := extractor.ExtractFromDirectory(absCwd)
		if err != nil {
			return fmt.Errorf("failed to extract documentation commands: %w", err)
		}
		commands = extracted
	}

	var activeRunner runner.Runner
	if flagExecute && !flagDryRun {
		activeRunner = runner.NewDockerRunner()
	} else {
		activeRunner = runner.NewDryRunner()
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Docker.Timeout)
	defer cancel()

	report, err := activeRunner.Run(ctx, absCwd, cfg, commands)
	if err != nil {
		return err
	}

	var rep reporter.Reporter
	if flagJSON {
		rep = reporter.NewJSONReporter(true)
	} else {
		rep = reporter.NewTextReporter(absCwd)
	}

	outStr, err := rep.Format(report)
	if err != nil {
		return fmt.Errorf("failed to format report: %w", err)
	}

	fmt.Print(outStr)

	if !report.Success {
		os.Exit(1)
	}

	return nil
}
