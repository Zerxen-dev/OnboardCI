package runner

import (
	"context"
	"time"

	"github.com/Zerxen-dev/onboardci/internal/config"
	"github.com/Zerxen-dev/onboardci/internal/extractor"
)

// MockStepResponse allows tests to mock step execution results.
type MockStepResponse struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// MockRunner simulates step execution without Docker.
type MockRunner struct {
	Responses map[string]MockStepResponse
}

func NewMockRunner(responses map[string]MockStepResponse) *MockRunner {
	if responses == nil {
		responses = make(map[string]MockStepResponse)
	}
	return &MockRunner{Responses: responses}
}

func (m *MockRunner) Run(ctx context.Context, cwd string, cfg *config.Config, commands []extractor.ExtractedCommand) (*ExecutionReport, error) {
	report := &ExecutionReport{
		Success:          true,
		TotalCommands:    len(commands),
		ExecutedInDocker: true,
		Steps:            make([]CommandResult, 0, len(commands)),
	}

	for _, cmd := range commands {
		resp, exists := m.Responses[cmd.Command]
		exitCode := 0
		stdout := ""
		stderr := ""
		if exists {
			exitCode = resp.ExitCode
			stdout = resp.Stdout
			stderr = resp.Stderr
		}

		res := CommandResult{
			Command:    cmd.Command,
			StepType:   string(cmd.Type),
			Source:     cmd.SourceFile,
			SourceFile: cmd.SourceFile,
			Line:       cmd.LineNumber,
			LineNumber: cmd.LineNumber,
			ExitCode:   exitCode,
			Status:     statusForExitCode(exitCode),
			Stdout:     stdout,
			Stderr:     stderr,
			Duration:   10 * time.Millisecond,
		}

		if exitCode != 0 {
			reason, suggestion := DiagnoseFailure(stderr+"\n"+stdout, exitCode)
			res.FailureReason = reason
			res.Reason = reason
			res.Suggestion = suggestion

			report.Success = false
			report.FailedStep = &res
			report.Steps = append(report.Steps, res)
			break
		}

		report.PassedCount++
		report.Steps = append(report.Steps, res)
	}

	return report, nil
}
