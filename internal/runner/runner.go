package runner

import (
	"context"
	"time"

	"github.com/Zerxen-dev/onboardci/internal/config"
	"github.com/Zerxen-dev/onboardci/internal/extractor"
)

// CommandResult records the outcome of a single step.
type CommandResult struct {
	Command       string        `json:"command"`
	StepType      string        `json:"step_type"`
	Source        string        `json:"source,omitempty"`
	SourceFile    string        `json:"source_file"`
	Line          int           `json:"line,omitempty"`
	LineNumber    int           `json:"line_number"`
	ExitCode      int           `json:"exit_code"`
	Status        string        `json:"status,omitempty"`
	Stdout        string        `json:"stdout,omitempty"`
	Stderr        string        `json:"stderr,omitempty"`
	Duration      time.Duration `json:"duration_ms"`
	Reason        string        `json:"reason,omitempty"`
	FailureReason string        `json:"failure_reason,omitempty"`
	Suggestion    string        `json:"suggestion,omitempty"`
}

// ExecutionReport summarizes all step outcomes.
type ExecutionReport struct {
	Success          bool            `json:"success"`
	TotalCommands    int             `json:"total_commands"`
	PassedCount      int             `json:"passed_count"`
	FailedStep       *CommandResult  `json:"failed_step,omitempty"`
	Steps            []CommandResult `json:"steps"`
	ExecutedInDocker bool            `json:"executed_in_docker"`
}

// statusForExitCode maps an exit code to a standardized status string.
func statusForExitCode(code int) string {
	if code == 0 {
		return "passed"
	}
	return "failed"
}

// Runner defines the interface for running extracted commands.
type Runner interface {
	Run(ctx context.Context, cwd string, cfg *config.Config, commands []extractor.ExtractedCommand) (*ExecutionReport, error)
}
