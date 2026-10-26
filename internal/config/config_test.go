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
