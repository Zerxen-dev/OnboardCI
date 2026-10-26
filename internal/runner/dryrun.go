package runner

import (
	"context"

	"github.com/Zerxen-dev/onboardci/internal/config"
	"github.com/Zerxen-dev/onboardci/internal/extractor"
)

// DryRunner formats extracted commands without container execution.
type DryRunner struct{}

func NewDryRunner() *DryRunner {
	return &DryRunner{}
}

func (r *DryRunner) Run(ctx context.Context, cwd string, cfg *config.Config, commands []extractor.ExtractedCommand) (*ExecutionReport, error) {
	report := &ExecutionReport{
		Success:          true,
		TotalCommands:    len(commands),
		PassedCount:      len(commands),
		ExecutedInDocker: false,
		Steps:            make([]CommandResult, 0, len(commands)),
	}

	for _, cmd := range commands {
		report.Steps = append(report.Steps, CommandResult{
			Command:    cmd.Command,
			StepType:   string(cmd.Type),
			Source:     cmd.SourceFile,
			SourceFile: cmd.SourceFile,
			Line:       cmd.LineNumber,
			LineNumber: cmd.LineNumber,
			ExitCode:   0,
			Status:     "passed",
		})
	}

	return report, nil
}
