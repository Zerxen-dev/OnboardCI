package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Zerxen-dev/onboardci/internal/config"
	"github.com/Zerxen-dev/onboardci/internal/extractor"
	"github.com/Zerxen-dev/onboardci/internal/runner"
)

func TestDryRunner(t *testing.T) {
	dry := runner.NewDryRunner()
	cfg := config.DefaultConfig()

	commands := []extractor.ExtractedCommand{
		{
			SourceFile: "README.md",
			LineNumber: 10,
			Type:       extractor.StepTypeSetup,
			Command:    "npm ci",
		},
		{
			SourceFile: "README.md",
			LineNumber: 20,
			Type:       extractor.StepTypeTest,
			Command:    "npm test",
		},
	}

	report, err := dry.Run(context.Background(), ".", cfg, commands)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.Success {
		t.Errorf("expected dry run to succeed")
	}
	if report.TotalCommands != 2 {
		t.Errorf("expected 2 total commands, got %d", report.TotalCommands)
	}
	if report.ExecutedInDocker {
		t.Errorf("expected ExecutedInDocker to be false for dry run")
	}
	if len(report.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(report.Steps))
	}
	if report.Steps[0].Command != "npm ci" || report.Steps[0].LineNumber != 10 {
		t.Errorf("step 0 mismatch: %+v", report.Steps[0])
	}
}

func TestDiagnoseFailure_MissingEnvVar(t *testing.T) {
	output := "Error: missing env var DATABASE_URL\n    at Object.<anonymous> (/app/index.js:12:15)"
	reason, suggestion := runner.DiagnoseFailure(output, 1)

	if reason != "missing env var DATABASE_URL" {
		t.Errorf("unexpected reason: %s", reason)
	}
	if suggestion != "document required env vars or provide .env.example" {
		t.Errorf("unexpected suggestion: %s", suggestion)
	}
}

func TestDiagnoseFailure_MissingEnvVar_WithEnvExample(t *testing.T) {
	tmpDir := t.TempDir()
	content := "# Configuration template\nDATABASE_URL=postgres://user:pass@localhost:5432/mydb\nPORT=3000\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .env.example: %v", err)
	}

	t.Chdir(tmpDir)

	output := "Error: missing env var DATABASE_URL\n    at Object.<anonymous> (/app/index.js:12:15)"
	reason, suggestion := runner.DiagnoseFailure(output, 1)

	if reason != "missing env var DATABASE_URL" {
		t.Errorf("unexpected reason: %s", reason)
	}
	expectedSuggestion := "Found .env.example with missing variable. Run cp .env.example .env and document required values."
	if suggestion != expectedSuggestion {
		t.Errorf("expected suggestion %q, got %q", expectedSuggestion, suggestion)
	}
}

func TestDiagnoseFailure_MissingEnvVar_WithEnvSample(t *testing.T) {
	tmpDir := t.TempDir()
	content := "export STRIPE_API_KEY=sk_test_12345\nexport WEBHOOK_SECRET=whsec_abc\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.sample"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .env.sample: %v", err)
	}

	t.Chdir(tmpDir)

	output := "TypeError: missing env var STRIPE_API_KEY"
	reason, suggestion := runner.DiagnoseFailure(output, 1)

	if reason != "missing env var STRIPE_API_KEY" {
		t.Errorf("unexpected reason: %s", reason)
	}
	expectedSuggestion := "Found .env.sample with missing variable. Run cp .env.sample .env and document required values."
	if suggestion != expectedSuggestion {
		t.Errorf("expected suggestion %q, got %q", expectedSuggestion, suggestion)
	}
}

func TestDiagnoseFailure_MissingEnvVar_NotInTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	content := "PORT=8080\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .env.example: %v", err)
	}

	t.Chdir(tmpDir)

	output := "Error: missing env var REDIS_URL"
	reason, suggestion := runner.DiagnoseFailure(output, 1)

	if reason != "missing env var REDIS_URL" {
		t.Errorf("unexpected reason: %s", reason)
	}
	expectedSuggestion := "document required env vars or provide .env.example"
	if suggestion != expectedSuggestion {
		t.Errorf("expected suggestion %q, got %q", expectedSuggestion, suggestion)
	}
}

func TestDiagnoseFailure_ProcessEnv(t *testing.T) {
	output := "TypeError: Cannot read properties of undefined (reading 'split')\nprocess.env.API_KEY is undefined"
	reason, suggestion := runner.DiagnoseFailure(output, 1)

	if reason != "missing env var API_KEY" {
		t.Errorf("unexpected reason: %s", reason)
	}
	if suggestion != "document required env vars or provide .env.example" {
		t.Errorf("unexpected suggestion: %s", suggestion)
	}
}

func TestDiagnoseFailure_MissingModule(t *testing.T) {
	output := "Error: Cannot find module 'express'\nRequire stack:"
	reason, suggestion := runner.DiagnoseFailure(output, 1)

	if reason != "missing dependency or tool 'express'" {
		t.Errorf("unexpected reason: %s", reason)
	}
	if suggestion != "document installation prerequisites or add to package manifest" {
		t.Errorf("unexpected suggestion: %s", suggestion)
	}
}

func TestMockRunner_FailureStopsExecution(t *testing.T) {
	responses := map[string]runner.MockStepResponse{
		"npm ci": {
			ExitCode: 0,
			Stdout:   "added 42 packages",
		},
		"npm test": {
			ExitCode: 1,
			Stderr:   "FAIL: missing env var DATABASE_URL",
		},
		"unreached cmd": {
			ExitCode: 0,
		},
	}

	mock := runner.NewMockRunner(responses)
	cfg := config.DefaultConfig()

	commands := []extractor.ExtractedCommand{
		{
			SourceFile: "README.md",
			LineNumber: 12,
			Type:       extractor.StepTypeSetup,
			Command:    "npm ci",
		},
		{
			SourceFile: "README.md",
			LineNumber: 42,
			Type:       extractor.StepTypeTest,
			Command:    "npm test",
		},
		{
			SourceFile: "README.md",
			LineNumber: 50,
			Type:       extractor.StepTypeTest,
			Command:    "unreached cmd",
		},
	}

	report, err := mock.Run(context.Background(), ".", cfg, commands)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Success {
		t.Errorf("expected report to be failure")
	}
	if report.PassedCount != 1 {
		t.Errorf("expected 1 passed count, got %d", report.PassedCount)
	}
	if report.FailedStep == nil {
		t.Fatalf("expected FailedStep to be set")
	}
	if report.FailedStep.LineNumber != 42 {
		t.Errorf("expected failed step line 42, got %d", report.FailedStep.LineNumber)
	}
	if report.FailedStep.FailureReason != "missing env var DATABASE_URL" {
		t.Errorf("expected missing env var DATABASE_URL, got %s", report.FailedStep.FailureReason)
	}
	// Check that 3rd command was not executed
	if len(report.Steps) != 2 {
		t.Errorf("expected exactly 2 executed steps, got %d", len(report.Steps))
	}
}
