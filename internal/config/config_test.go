package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Zerxen-dev/onboardci/internal/config"
)

func TestLoadConfig_Defaults(t *testing.T) {
	tempDir := t.TempDir()

	cfg, path, err := config.LoadConfig(tempDir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty path for non-existent config, got %s", path)
	}
	if cfg.Docker.Image != "node:20-alpine" {
		t.Errorf("expected default image node:20-alpine, got %s", cfg.Docker.Image)
	}
	if cfg.Docker.Timeout != 10*time.Minute {
		t.Errorf("expected default timeout 10m, got %v", cfg.Docker.Timeout)
	}
	if cfg.Docker.User != "1000:1000" {
		t.Errorf("expected default user 1000:1000, got %s", cfg.Docker.User)
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".onboardci.yml")
	content := `
setup:
  - npm ci
  - npm run build
test:
  - npm test
env:
  - DATABASE_URL=sqlite://test.db
  - NODE_ENV=test
docker:
  image: node:22-bookworm-slim
  timeout: 5m
  user: "1001:1001"
  cpus: "1.5"
  memory: "1g"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, foundPath, err := config.LoadConfig(tempDir, "")
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	if foundPath != configPath {
		t.Errorf("expected foundPath %s, got %s", configPath, foundPath)
	}
	if len(cfg.Setup) != 2 || cfg.Setup[0] != "npm ci" {
		t.Errorf("unexpected setup commands: %v", cfg.Setup)
	}
	if len(cfg.Test) != 1 || cfg.Test[0] != "npm test" {
		t.Errorf("unexpected test commands: %v", cfg.Test)
	}
	if len(cfg.Env) != 2 || cfg.Env[0] != "DATABASE_URL=sqlite://test.db" {
		t.Errorf("unexpected env: %v", cfg.Env)
	}
	if cfg.Docker.Image != "node:22-bookworm-slim" {
		t.Errorf("unexpected image: %s", cfg.Docker.Image)
	}
	if cfg.Docker.Timeout != 5*time.Minute {
		t.Errorf("unexpected timeout: %v", cfg.Docker.Timeout)
	}
	if cfg.Docker.CPUs != "1.5" {
		t.Errorf("unexpected cpus: %s", cfg.Docker.CPUs)
	}
	if cfg.Docker.Memory != "1g" {
		t.Errorf("unexpected memory: %s", cfg.Docker.Memory)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".onboardci.yml")
	content := `
setup: [invalid: yaml:
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	_, _, err := config.LoadConfig(tempDir, "")
	if err == nil {
		t.Fatalf("expected error for invalid yaml, got nil")
	}
}

func TestLoadConfig_SecurityValidation(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "malicious user command injection",
			yaml: "docker:\n  user: \"1000; rm -rf /\"\n",
			wantErr: true,
		},
		{
			name: "malicious user with spaces",
			yaml: "docker:\n  user: \"root /workspace\"\n",
			wantErr: true,
		},
		{
			name: "valid user:group format",
			yaml: "docker:\n  user: \"1000:1000\"\n",
			wantErr: false,
		},
		{
			name: "valid username:groupname",
			yaml: "docker:\n  user: \"node:node\"\n",
			wantErr: false,
		},
		{
			name: "invalid cpus with flag injection",
			yaml: "docker:\n  cpus: \"--privileged\"\n",
			wantErr: true,
		},
		{
			name: "valid numeric cpus",
			yaml: "docker:\n  cpus: \"2.5\"\n",
			wantErr: false,
		},
		{
			name: "invalid memory format",
			yaml: "docker:\n  memory: \"2gb; reboot\"\n",
			wantErr: true,
		},
		{
			name: "valid memory size",
			yaml: "docker:\n  memory: \"512m\"\n",
			wantErr: false,
		},
		{
			name: "malicious image with shell chars",
			yaml: "docker:\n  image: \"node:20; curl evil.com\"\n",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			path := filepath.Join(tmp, ".onboardci.yml")
			if err := os.WriteFile(path, []byte(tc.yaml), 0644); err != nil {
				t.Fatalf("failed to write yaml: %v", err)
			}
			_, _, err := config.LoadConfig(tmp, "")
			if (err != nil) != tc.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
