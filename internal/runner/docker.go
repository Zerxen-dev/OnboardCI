package runner

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"github.com/Zerxen-dev/onboardci/internal/config"
	"github.com/Zerxen-dev/onboardci/internal/extractor"
)

const maxLogBytes = 64 * 1024 // 64KB per step log cap

// DockerRunner manages execution inside isolated Docker containers.
type DockerRunner struct{}

func NewDockerRunner() *DockerRunner {
	return &DockerRunner{}
}

// CheckDockerAvailable verifies that docker CLI exists.
func CheckDockerAvailable() error {
	_, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("docker binary not found in PATH: %w", err)
	}
	return nil
}

func (r *DockerRunner) Run(ctx context.Context, cwd string, cfg *config.Config, commands []extractor.ExtractedCommand) (*ExecutionReport, error) {
	if err := CheckDockerAvailable(); err != nil {
		return nil, err
	}

	report := &ExecutionReport{
		Success:          true,
		TotalCommands:    len(commands),
		ExecutedInDocker: true,
		Steps:            make([]CommandResult, 0, len(commands)),
	}

	// Generate a unique container name
	containerName := fmt.Sprintf("onboardci-%d-%d", time.Now().UnixNano(), rand.Intn(10000))

	// 1. Create container
	createArgs := []string{
		"create",
		"--name", containerName,
		"--user", cfg.Docker.User,
		"--cpus", cfg.Docker.CPUs,
		"--memory", cfg.Docker.Memory,
		"--network", "bridge",
		"-w", "/workspace",
		cfg.Docker.Image,
		"tail", "-f", "/dev/null",
	}

	createCmd := exec.CommandContext(ctx, "docker", createArgs...)
	out, err := createCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker container: %s: %w", string(out), err)
	}

	// Ensure cleanup on return
	defer func() {
		rmCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = exec.CommandContext(rmCtx, "docker", "rm", "-f", containerName).Run()
	}()

	// 2. Start container
	startCmd := exec.CommandContext(ctx, "docker", "start", containerName)
	if out, err := startCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to start docker container: %s: %w", string(out), err)
	}

	// 3. Copy repository into /workspace (NO host bind mounts)
	// First ensure /workspace exists and is owned by non-root user
	initPermsCmd := exec.CommandContext(ctx, "docker", "exec", "--user", "0:0", containerName, "sh", "-c", "mkdir -p /workspace && chown -R "+cfg.Docker.User+" /workspace")
	_ = initPermsCmd.Run()

	cpCmd := exec.CommandContext(ctx, "docker", "cp", cwd+"/.", containerName+":/workspace/")
	if out, err := cpCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to copy repository to container: %s: %w", string(out), err)
	}

	// Chown workspace again after copy so user 1000 has full write access
	chownCmd := exec.CommandContext(ctx, "docker", "exec", "--user", "0:0", containerName, "chown", "-R", cfg.Docker.User, "/workspace")
	_ = chownCmd.Run()

	// 4. Sequentially execute commands
	for _, cmd := range commands {
		stepStart := time.Now()

		var execArgs []string
		execArgs = append(execArgs, "exec", "-i", "--user", cfg.Docker.User, "-w", "/workspace")

		// Pass environment variables from config
		for _, envVar := range cfg.Env {
			execArgs = append(execArgs, "-e", envVar)
		}

		execArgs = append(execArgs, containerName, "sh", "-c", cmd.Command)

		stepCmd := exec.CommandContext(ctx, "docker", execArgs...)

		var stdoutBuf, stderrBuf bytes.Buffer
		stepCmd.Stdout = &stdoutBuf
		stepCmd.Stderr = &stderrBuf

		err := stepCmd.Run()
		duration := time.Since(stepStart)

		stdout := truncateOutput(stdoutBuf.String(), maxLogBytes)
		stderr := truncateOutput(stderrBuf.String(), maxLogBytes)

		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 1
			}
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
			Duration:   duration,
		}

		if exitCode != 0 {
			combined := stderr + "\n" + stdout
			reason, suggestion := DiagnoseFailure(combined, exitCode)
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

func truncateOutput(s string, limit int) string {
	if len(s) <= limit {
		return strings.TrimSpace(s)
	}
	half := limit / 2
	return strings.TrimSpace(s[:half]) + "\n...[output truncated]...\n" + strings.TrimSpace(s[len(s)-half:])
}
