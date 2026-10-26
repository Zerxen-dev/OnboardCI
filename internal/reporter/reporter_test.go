package reporter_test

import (
	"strings"
	"testing"

	"github.com/Zerxen-dev/onboardci/internal/reporter"
	"github.com/Zerxen-dev/onboardci/internal/runner"
)

func TestTextReporter_FailureFormat(t *testing.T) {
	rep := reporter.NewTextReporter("")

	report := &runner.ExecutionReport{
		Success:          false,
		TotalCommands:    2,
		ExecutedInDocker: true,
		FailedStep: &runner.CommandResult{
			Command:       "npm test",
			SourceFile:    "README.md",
			LineNumber:    42,
			ExitCode:      1,
			FailureReason: "missing env var DATABASE_URL",
			Suggestion:    "document required env vars or provide .env.example",
		},
	}

	out, err := rep.Format(report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParts := []string{
		"FAIL: README.md line 42",
		"Command: npm test",
		"Reason: missing env var DATABASE_URL",
		"Suggestion: document required env vars or provide .env.example",
	}

	for _, part := range expectedParts {
		if !strings.Contains(out, part) {
			t.Errorf("output missing expected part: %q\nGot:\n%s", part, out)
		}
	}
}

func TestJSONReporter(t *testing.T) {
	rep := reporter.NewJSONReporter(true)

	report := &runner.ExecutionReport{
		Success:       true,
		TotalCommands: 1,
	}

	out, err := rep.Format(report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, `"success": true`) {
		t.Errorf("json missing success flag: %s", out)
	}
}

func TestGitHubPRReporter(t *testing.T) {
	rep := reporter.NewGitHubPRReporter("")

	report := &runner.ExecutionReport{
		Success:          false,
		TotalCommands:    1,
		ExecutedInDocker: true,
		FailedStep: &runner.CommandResult{
			Command:       "npm test",
			SourceFile:    "README.md",
			LineNumber:    42,
			ExitCode:      1,
			FailureReason: "missing env var DATABASE_URL",
			Suggestion:    "document required env vars or provide .env.example",
		},
		Steps: []runner.CommandResult{
			{
				Command:  "npm test",
				StepType: "test",
				ExitCode: 1,
			},
		},
	}

	out, err := rep.Format(report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "First-Time Contributor Blocked") {
		t.Errorf("github comment missing blocked title")
	}
	if !strings.Contains(out, "README.md") {
		t.Errorf("github comment missing file name")
	}
}
