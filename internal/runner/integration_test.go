//go:build integration

package runner_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Zerxen-dev/onboardci/internal/config"
	"github.com/Zerxen-dev/onboardci/internal/extractor"
	"github.com/Zerxen-dev/onboardci/internal/reporter"
	"github.com/Zerxen-dev/onboardci/internal/runner"
)

func TestIntegration_BrokenRepoFailure(t *testing.T) {
	fixtureDir, err := filepath.Abs("../../testdata/fixtures/broken-env")
	if err != nil {
		t.Fatalf("failed to resolve fixture directory: %v", err)
	}
	if _, err := os.Stat(fixtureDir); err != nil {
		// Fallback to broken-repo if broken-env does not exist
		fixtureDir, err = filepath.Abs("../../testdata/fixtures/broken-repo")
		if err != nil {
			t.Fatalf("failed to resolve fallback fixture directory: %v", err)
		}
	}

	cfg, _, err := config.LoadConfig(fixtureDir, "")
	if err != nil {
		t.Fatalf("failed to load fixture config: %v", err)
	}

	commands, err := extractor.ExtractFromDirectory(fixtureDir)
	if err != nil {
		t.Fatalf("failed to extract commands: %v", err)
	}

	if len(commands) == 0 {
		t.Fatalf("expected extracted commands from broken fixture")
	}

	// Determine if Docker daemon is accessible
	dockerAvailable := false
	checkCmd := exec.Command("docker", "info")
	if err := checkCmd.Run(); err == nil {
		dockerAvailable = true
	}

	var r runner.Runner
	if dockerAvailable {
		t.Log("Docker daemon active: executing in real Docker sandbox")
		r = runner.NewDockerRunner()
	} else {
		// Guard: In CI or when explicitly required, fail loudly if Docker is missing so CI truth is visible
		if os.Getenv("CI") != "" || os.Getenv("REQUIRE_DOCKER") != "" {
			t.Fatalf("Docker daemon is required for integration tests in CI/strict mode, but 'docker info' failed")
		}

		t.Log("Docker daemon unavailable: executing through simulated runner engine (non-CI)")
		r = runner.NewMockRunner(map[string]runner.MockStepResponse{
			"npm install": {
				ExitCode: 0,
				Stdout:   "added 1 package in 0.2s",
			},
			"npm test": {
				ExitCode: 1,
				Stderr:   "FAIL: missing env var DATABASE_URL",
			},
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	report, err := r.Run(ctx, fixtureDir, cfg, commands)
	if err != nil {
		t.Fatalf("runner execution error: %v", err)
	}

	// 1. Assert failure
	if report.Success {
		t.Errorf("expected broken fixture execution to fail")
	}

	// 2. Assert failed step exists
	if report.FailedStep == nil {
		t.Fatalf("expected FailedStep to be populated")
	}

	// 3. Assert failure line number
	if report.FailedStep.LineNumber != 42 {
		t.Errorf("expected failure on line 42, got %d", report.FailedStep.LineNumber)
	}

	// 4. Assert failure command
	if report.FailedStep.Command != "npm test" {
		t.Errorf("expected command 'npm test', got %s", report.FailedStep.Command)
	}

	// 5. Assert diagnosed reason and suggestion
	if report.FailedStep.FailureReason != "missing env var DATABASE_URL" {
		t.Errorf("expected reason 'missing env var DATABASE_URL', got %s", report.FailedStep.FailureReason)
	}
	if report.FailedStep.Suggestion != "document required env vars or provide .env.example" {
		t.Errorf("expected suggestion 'document required env vars or provide .env.example', got %s", report.FailedStep.Suggestion)
	}

	// 6. Test text reporter output matches prompt requirements
	rep := reporter.NewTextReporter(fixtureDir)
	textReport, err := rep.Format(report)
	if err != nil {
		t.Fatalf("formatting error: %v", err)
	}

	expectedSnippet := "FAIL: README.md line 42\nCommand: npm test\nReason: missing env var DATABASE_URL\nSuggestion: document required env vars or provide .env.example"
	if !strings.Contains(textReport, expectedSnippet) {
		t.Errorf("formatted report does not match expected output format.\nGot:\n%s\nExpected to contain:\n%s", textReport, expectedSnippet)
	}
}
