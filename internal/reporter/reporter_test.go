package reporter_test

import (
	"encoding/json"
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

func TestJSONReporter_StandardizedSchema(t *testing.T) {
	rep := reporter.NewJSONReporter(false)

	report := &runner.ExecutionReport{
		Success:          false,
		TotalCommands:    2,
		PassedCount:      1,
		ExecutedInDocker: true,
		FailedStep: &runner.CommandResult{
			Command:       "npm test",
			StepType:      "test",
			Source:        "README.md",
			SourceFile:    "README.md",
			Line:          42,
			LineNumber:    42,
			ExitCode:      1,
			Status:        "failed",
			Reason:        "missing env var DATABASE_URL",
			FailureReason: "missing env var DATABASE_URL",
			Suggestion:    "document required env vars or provide .env.example",
		},
		Steps: []runner.CommandResult{
			{
				Command:    "npm ci",
				StepType:   "setup",
				Source:     "README.md",
				SourceFile: "README.md",
				Line:       10,
				LineNumber: 10,
				ExitCode:   0,
				Status:     "passed",
			},
			{
				Command:       "npm test",
				StepType:      "test",
				Source:        "README.md",
				SourceFile:    "README.md",
				Line:          42,
				LineNumber:    42,
				ExitCode:      1,
				Status:        "failed",
				Reason:        "missing env var DATABASE_URL",
				FailureReason: "missing env var DATABASE_URL",
				Suggestion:    "document required env vars or provide .env.example",
			},
		},
	}

	out, err := rep.Format(report)
	if err != nil {
		t.Fatalf("unexpected format error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if parsed["success"] != false {
		t.Errorf("expected success=false, got %v", parsed["success"])
	}
	if parsed["total_commands"] != float64(2) {
		t.Errorf("expected total_commands=2, got %v", parsed["total_commands"])
	}
	if parsed["passed_count"] != float64(1) {
		t.Errorf("expected passed_count=1, got %v", parsed["passed_count"])
	}
	if parsed["executed_in_docker"] != true {
		t.Errorf("expected executed_in_docker=true, got %v", parsed["executed_in_docker"])
	}

	failedStep, ok := parsed["failed_step"].(map[string]interface{})
	if !ok {
		t.Fatalf("failed_step missing or invalid: %v", parsed["failed_step"])
	}
	if failedStep["status"] != "failed" {
		t.Errorf("failed_step.status mismatch: %v", failedStep["status"])
	}
	if failedStep["source"] != "README.md" || failedStep["line"] != float64(42) {
		t.Errorf("failed_step source/line mismatch: %v:%v", failedStep["source"], failedStep["line"])
	}
	if failedStep["reason"] != "missing env var DATABASE_URL" {
		t.Errorf("failed_step.reason mismatch: %v", failedStep["reason"])
	}

	steps, ok := parsed["steps"].([]interface{})
	if !ok || len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %v", len(steps))
	}
	step0 := steps[0].(map[string]interface{})
	if step0["status"] != "passed" || step0["source"] != "README.md" || step0["line"] != float64(10) {
		t.Errorf("step 0 fields mismatch: %+v", step0)
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
