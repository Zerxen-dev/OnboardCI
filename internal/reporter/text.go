package reporter

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Zerxen-dev/onboardci/internal/runner"
)

// TextReporter formats human-readable CLI outputs.
type TextReporter struct {
	BaseDir string
}

func NewTextReporter(baseDir string) *TextReporter {
	return &TextReporter{BaseDir: baseDir}
}

func (r *TextReporter) Format(report *runner.ExecutionReport) (string, error) {
	var sb strings.Builder

	// Dry-run mode
	if !report.ExecutedInDocker {
		sb.WriteString("OnboardCI — Dry Run (Listing Extracted Commands)\n\n")
		if len(report.Steps) == 0 {
			sb.WriteString("No setup or test commands found in documentation or .onboardci.yml.\n")
			return sb.String(), nil
		}

		for i, step := range report.Steps {
			relFile := step.SourceFile
			if r.BaseDir != "" {
				if rel, err := filepath.Rel(r.BaseDir, step.SourceFile); err == nil && !strings.HasPrefix(rel, "..") {
					relFile = rel
				}
			}
			if step.LineNumber > 0 {
				sb.WriteString(fmt.Sprintf("[%d] %s (%s line %d):\n", i+1, step.StepType, relFile, step.LineNumber))
			} else {
				sb.WriteString(fmt.Sprintf("[%d] %s (%s):\n", i+1, step.StepType, relFile))
			}
			sb.WriteString(fmt.Sprintf("    %s\n", step.Command))
		}

		sb.WriteString(fmt.Sprintf("\nTotal: %d commands extracted.\n", len(report.Steps)))
		sb.WriteString("Run with '--execute' to verify them inside a clean Docker sandbox.\n")
		return sb.String(), nil
	}

	// Execution mode - Failure
	if !report.Success && report.FailedStep != nil {
		failed := report.FailedStep
		relFile := failed.SourceFile
		if r.BaseDir != "" {
			if rel, err := filepath.Rel(r.BaseDir, failed.SourceFile); err == nil && !strings.HasPrefix(rel, "..") {
				relFile = rel
			}
		}

		if failed.LineNumber > 0 {
			sb.WriteString(fmt.Sprintf("FAIL: %s line %d\n", relFile, failed.LineNumber))
		} else {
			sb.WriteString(fmt.Sprintf("FAIL: %s\n", relFile))
		}
		sb.WriteString(fmt.Sprintf("Command: %s\n", failed.Command))
		sb.WriteString(fmt.Sprintf("Reason: %s\n", failed.FailureReason))
		sb.WriteString(fmt.Sprintf("Suggestion: %s\n", failed.Suggestion))

		if failed.Stderr != "" {
			sb.WriteString(fmt.Sprintf("\nCaptured Error Output:\n%s\n", failed.Stderr))
		}
		return sb.String(), nil
	}

	// Execution mode - Pass
	sb.WriteString(fmt.Sprintf("PASS: Verified %d onboarding commands successfully in container sandbox.\n", report.TotalCommands))
	for i, step := range report.Steps {
		sb.WriteString(fmt.Sprintf("  [%d] %s (%s) — ok\n", i+1, step.Command, step.StepType))
	}
	return sb.String(), nil
}
