package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Zerxen-dev/onboardci/internal/cli"
)

func TestDetectProject_Node(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "package.json"), []byte(`{"name":"test-app"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "package-lock.json"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	preset := cli.DetectProject(tmp)
	if preset.Language != "Node.js" {
		t.Errorf("expected Node.js, got %s", preset.Language)
	}
	if preset.Image != "node:20-alpine" {
		t.Errorf("expected node:20-alpine, got %s", preset.Image)
	}
	if len(preset.Setup) == 0 || preset.Setup[0] != "npm ci" {
		t.Errorf("expected setup npm ci, got %v", preset.Setup)
	}
	if len(preset.Test) == 0 || preset.Test[0] != "npm test" {
		t.Errorf("expected test npm test, got %v", preset.Test)
	}
}

func TestDetectProject_Go(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\ngo 1.23\n"), 0644); err != nil {
		t.Fatal(err)
	}

	preset := cli.DetectProject(tmp)
	if preset.Language != "Go" {
		t.Errorf("expected Go, got %s", preset.Language)
	}
	if preset.Image != "golang:1.23-alpine" {
		t.Errorf("expected golang:1.23-alpine, got %s", preset.Image)
	}
	if len(preset.Setup) == 0 || preset.Setup[0] != "go mod download" {
		t.Errorf("expected setup go mod download, got %v", preset.Setup)
	}
	if len(preset.Test) == 0 || preset.Test[0] != "go test ./..." {
		t.Errorf("expected test go test ./..., got %v", preset.Test)
	}
}

func TestDetectProject_Python(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "requirements.txt"), []byte("pytest>=8.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	preset := cli.DetectProject(tmp)
	if preset.Language != "Python" {
		t.Errorf("expected Python, got %s", preset.Language)
	}
	if preset.Image != "python:3.12-slim" {
		t.Errorf("expected python:3.12-slim, got %s", preset.Image)
	}
	if len(preset.Setup) == 0 || preset.Setup[0] != "pip install -r requirements.txt" {
		t.Errorf("expected pip install, got %v", preset.Setup)
	}
	if len(preset.Test) == 0 || preset.Test[0] != "pytest" {
		t.Errorf("expected pytest, got %v", preset.Test)
	}
}

func TestDetectProject_Rust(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "Cargo.toml"), []byte("[package]\nname = \"app\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	preset := cli.DetectProject(tmp)
	if preset.Language != "Rust" {
		t.Errorf("expected Rust, got %s", preset.Language)
	}
	if preset.Image != "rust:alpine" {
		t.Errorf("expected rust:alpine, got %s", preset.Image)
	}
	if len(preset.Setup) == 0 || preset.Setup[0] != "cargo build" {
		t.Errorf("expected cargo build, got %v", preset.Setup)
	}
	if len(preset.Test) == 0 || preset.Test[0] != "cargo test" {
		t.Errorf("expected cargo test, got %v", preset.Test)
	}
}

func TestDetectProject_Makefile(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "Makefile"), []byte("all:\n\t@echo ok\n"), 0644); err != nil {
		t.Fatal(err)
	}

	preset := cli.DetectProject(tmp)
	if preset.Language != "Makefile" {
		t.Errorf("expected Makefile, got %s", preset.Language)
	}
	if preset.Image != "ubuntu:22.04" {
		t.Errorf("expected ubuntu:22.04, got %s", preset.Image)
	}
}

func TestInitCmd_Execution(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/myproject\ngo 1.23\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. Initial run should create .onboardci.yml
	cmd := cli.RootCmd
	cmd.SetArgs([]string{"init", "--cwd", tmp})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init execution failed: %v", err)
	}

	cfgPath := filepath.Join(tmp, ".onboardci.yml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf(".onboardci.yml was not created: %v", err)
	}
	if !strings.Contains(string(data), "golang:1.23-alpine") {
		t.Errorf("expected golang image in config, got:\n%s", string(data))
	}

	// 2. Second run without --force should fail
	cmd.SetArgs([]string{"init", "--cwd", tmp})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error when .onboardci.yml already exists without --force")
	}

	// 3. Second run with --force should succeed
	cmd.SetArgs([]string{"init", "--cwd", tmp, "--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected --force to succeed, got: %v", err)
	}

	// 4. Dry run should not write to custom non-existent directory
	dryDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dryDir, "package.json"), []byte(`{"name":"dry"}`), 0644); err != nil {
		t.Fatal(err)
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"init", "--cwd", dryDir, "--dry-run"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dryDir, ".onboardci.yml")); err == nil {
		t.Errorf("expected .onboardci.yml not to be written during dry-run")
	}
}
