package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	flagInitForce  bool
	flagInitDryRun bool
)

// ProjectPreset encapsulates suggested configuration based on repository contents.
type ProjectPreset struct {
	Language string   `json:"language"`
	Image    string   `json:"image"`
	Setup    []string `json:"setup"`
	Test     []string `json:"test"`
	Env      []string `json:"env"`
}

// DetectProject inspects the directory for manifests and returns the matching preset.
func DetectProject(dir string) *ProjectPreset {
	// 1. Node.js
	if fileExists(filepath.Join(dir, "package.json")) {
		setupCmd := "npm install"
		if fileExists(filepath.Join(dir, "package-lock.json")) {
			setupCmd = "npm ci"
		} else if fileExists(filepath.Join(dir, "yarn.lock")) {
			setupCmd = "yarn install --frozen-lockfile"
		} else if fileExists(filepath.Join(dir, "pnpm-lock.yaml")) {
			setupCmd = "pnpm install --frozen-lockfile"
		} else if fileExists(filepath.Join(dir, "bun.lockb")) || fileExists(filepath.Join(dir, "bun.lock")) {
			setupCmd = "bun install"
		}

		return &ProjectPreset{
			Language: "Node.js",
			Image:    "node:20-alpine",
			Setup:    []string{setupCmd},
			Test:     []string{"npm test"},
			Env:      []string{},
		}
	}

	// 2. Go
	if fileExists(filepath.Join(dir, "go.mod")) {
		return &ProjectPreset{
			Language: "Go",
			Image:    "golang:1.23-alpine",
			Setup:    []string{"go mod download"},
			Test:     []string{"go test ./..."},
			Env:      []string{},
		}
	}

	// 3. Python
	if fileExists(filepath.Join(dir, "requirements.txt")) || fileExists(filepath.Join(dir, "pyproject.toml")) || fileExists(filepath.Join(dir, "Pipfile")) {
		setupCmd := "pip install -r requirements.txt"
		testCmd := "pytest"

		if fileExists(filepath.Join(dir, "poetry.lock")) {
			setupCmd = "poetry install"
			testCmd = "poetry run pytest"
		} else if fileExists(filepath.Join(dir, "Pipfile")) {
			setupCmd = "pipenv install --dev"
			testCmd = "pipenv run pytest"
		}

		return &ProjectPreset{
			Language: "Python",
			Image:    "python:3.12-slim",
			Setup:    []string{setupCmd},
			Test:     []string{testCmd},
			Env:      []string{},
		}
	}

	// 4. Rust
	if fileExists(filepath.Join(dir, "Cargo.toml")) {
		return &ProjectPreset{
			Language: "Rust",
			Image:    "rust:alpine",
			Setup:    []string{"cargo build"},
			Test:     []string{"cargo test"},
			Env:      []string{},
		}
	}

	// 5. Makefile
	if fileExists(filepath.Join(dir, "Makefile")) || fileExists(filepath.Join(dir, "makefile")) {
		return &ProjectPreset{
			Language: "Makefile",
			Image:    "ubuntu:22.04",
			Setup:    []string{"make build"},
			Test:     []string{"make test"},
			Env:      []string{},
		}
	}

	// Default fallback
	return &ProjectPreset{
		Language: "Generic",
		Image:    "node:20-alpine",
		Setup:    []string{},
		Test:     []string{},
		Env:      []string{},
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// GenerateConfigYAML produces a structured, commented .onboardci.yml string.
func GenerateConfigYAML(preset *ProjectPreset) ([]byte, error) {
	type DockerOut struct {
		Image   string `yaml:"image"`
		Timeout string `yaml:"timeout"`
		User    string `yaml:"user"`
		CPUs    string `yaml:"cpus"`
		Memory  string `yaml:"memory"`
	}

	type ConfigOut struct {
		Setup  []string  `yaml:"setup"`
		Test   []string  `yaml:"test"`
		Env    []string  `yaml:"env"`
		Docker DockerOut `yaml:"docker"`
	}

	out := ConfigOut{
		Setup: preset.Setup,
		Test:  preset.Test,
		Env:   preset.Env,
		Docker: DockerOut{
			Image:   preset.Image,
			Timeout: "10m",
			User:    "1000:1000",
			CPUs:    "2.0",
			Memory:  "2g",
		},
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("# OnboardCI Configuration File\n# Generated for %s project\n\n", preset.Language))

	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(&out); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scan repository and generate a starter .onboardci.yml configuration",
	Long: `Inspects the target repository for project manifests (package.json, go.mod,
Cargo.toml, requirements.txt, Makefile) and generates a starter .onboardci.yml file.`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVarP(&flagInitForce, "force", "f", false, "Overwrite existing .onboardci.yml without prompting")
	initCmd.Flags().BoolVar(&flagInitDryRun, "dry-run", false, "Preview generated configuration on stdout without writing")
}

func runInit(cmd *cobra.Command, args []string) error {
	absCwd, err := filepath.Abs(flagCwd)
	if err != nil {
		return fmt.Errorf("invalid working directory %q: %w", flagCwd, err)
	}

	targetPath := filepath.Join(absCwd, ".onboardci.yml")
	altPath := filepath.Join(absCwd, ".onboardci.yaml")

	if !flagInitDryRun && !flagInitForce {
		if fileExists(targetPath) || fileExists(altPath) {
			return fmt.Errorf(".onboardci.yml already exists (use --force to overwrite)")
		}
	}

	preset := DetectProject(absCwd)
	data, err := GenerateConfigYAML(preset)
	if err != nil {
		return fmt.Errorf("failed to generate configuration: %w", err)
	}

	if flagInitDryRun {
		cmd.Println(string(data))
		return nil
	}

	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", targetPath, err)
	}

	cmd.Printf("✓ Created %s for %s project\n", targetPath, preset.Language)
	cmd.Println("Run 'onboardci check' to verify onboarding steps.")
	return nil
}
