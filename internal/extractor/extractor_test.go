package extractor_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Zerxen-dev/onboardci/internal/extractor"
)

func TestExtractFromMarkdown(t *testing.T) {
	content := `# Sample Project

Welcome to Sample Project.

## Installation

First, install dependencies:
` + "```bash\n" +
		"# Install packages\n" +
		"npm ci\n" +
		"npm run build\n" +
		"```\n" +
		"\n" +
		"## Testing\n" +
		"\n" +
		"Run the test suite:\n" +
		"```sh\n" +
		"$ npm test\n" +
		"```\n" +
		"\n" +
		"Unrelated block:\n" +
		"```json\n" +
		"{\"key\": \"value\"}\n" +
		"```\n"

	tempFile := filepath.Join(t.TempDir(), "README.md")
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	cmds, err := extractor.ExtractFromMarkdown(tempFile, []byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cmds) != 3 {
		t.Fatalf("expected 3 commands, got %d: %+v", len(cmds), cmds)
	}

	// 1: npm ci (line 10)
	if cmds[0].Command != "npm ci" || cmds[0].Type != extractor.StepTypeSetup || cmds[0].LineNumber != 10 {
		t.Errorf("cmd 0 mismatch: %+v", cmds[0])
	}
	// 2: npm run build (line 11)
	if cmds[0].SourceFile != tempFile {
		t.Errorf("expected source file %s, got %s", tempFile, cmds[0].SourceFile)
	}
	if cmds[1].Command != "npm run build" || cmds[1].Type != extractor.StepTypeSetup || cmds[1].LineNumber != 11 {
		t.Errorf("cmd 1 mismatch: %+v", cmds[1])
	}
	// 3: npm test (line 18, leading $ stripped)
	if cmds[2].Command != "npm test" || cmds[2].Type != extractor.StepTypeTest || cmds[2].LineNumber != 18 {
		t.Errorf("cmd 2 mismatch: %+v", cmds[2])
	}
}

func TestExtractFromDirectory_Order(t *testing.T) {
	tempDir := t.TempDir()

	readmeContent := `
# Project
` + "```bash\n" +
		"pip install -r requirements.txt\n" +
		"pytest\n" +
		"```\n"

	contributingContent := `
# Contributing
` + "```bash\n" +
		"poetry install\n" +
		"poetry run pytest\n" +
		"```\n"

	_ = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte(readmeContent), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "CONTRIBUTING.md"), []byte(contributingContent), 0644)

	cmds, err := extractor.ExtractFromDirectory(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cmds) != 4 {
		t.Fatalf("expected 4 commands from README and CONTRIBUTING, got %d", len(cmds))
	}
}

func TestExtractFromDirectory_PythonManifestFallback(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "requirements.txt"), []byte("pytest>=8.0.0\nrequests>=2.31.0\n"), 0644)

	cmds, err := extractor.ExtractFromDirectory(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cmds) != 2 {
		t.Fatalf("expected 2 synthesized commands, got %d", len(cmds))
	}
	if cmds[0].Command != "pip install -r requirements.txt" || cmds[0].Type != extractor.StepTypeSetup {
		t.Errorf("unexpected setup command: %+v", cmds[0])
	}
	if cmds[1].Command != "pytest" || cmds[1].Type != extractor.StepTypeTest {
		t.Errorf("unexpected test command: %+v", cmds[1])
	}
}

func TestExtractFromDirectory_PoetryManifestFallback(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "pyproject.toml"), []byte("[tool.poetry]\nname = \"pkg\"\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "poetry.lock"), []byte(""), 0644)

	cmds, err := extractor.ExtractFromDirectory(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cmds) != 2 {
		t.Fatalf("expected 2 synthesized commands, got %d", len(cmds))
	}
	if cmds[0].Command != "poetry install" || cmds[0].Type != extractor.StepTypeSetup {
		t.Errorf("unexpected setup command: %+v", cmds[0])
	}
	if cmds[1].Command != "pytest" || cmds[1].Type != extractor.StepTypeTest {
		t.Errorf("unexpected test command: %+v", cmds[1])
	}
}

func TestClassifyCommand_PythonCommands(t *testing.T) {
	testCases := []struct {
		cmd      string
		wantType extractor.StepType
		wantOk   bool
	}{
		{"pytest", extractor.StepTypeTest, true},
		{"python3 -m unittest discover", extractor.StepTypeTest, true},
		{"poetry run pytest", extractor.StepTypeTest, true},
		{"tox", extractor.StepTypeTest, true},
		{"pip install -r requirements.txt", extractor.StepTypeSetup, true},
		{"uv pip install -e .", extractor.StepTypeSetup, true},
		{"poetry install", extractor.StepTypeSetup, true},
		{"pipenv install", extractor.StepTypeSetup, true},
	}

	for _, tc := range testCases {
		stepType, ok := extractor.ClassifyCommand(tc.cmd)
		if ok != tc.wantOk || stepType != tc.wantType {
			t.Errorf("ClassifyCommand(%q) = (%v, %v), want (%v, %v)", tc.cmd, stepType, ok, tc.wantType, tc.wantOk)
		}
	}
}

func TestDetectPackageNames(t *testing.T) {
	tmp := t.TempDir()

	// 1. package.json with scoped name
	pkgJSON := `{"name": "@types/webpack"}`
	_ = os.WriteFile(filepath.Join(tmp, "package.json"), []byte(pkgJSON), 0644)

	names := extractor.DetectPackageNames(tmp)
	if len(names) != 2 || names[0] != "@types/webpack" || names[1] != "webpack" {
		t.Errorf("unexpected package names from package.json: %v", names)
	}

	// 2. go.mod
	tmpGo := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpGo, "go.mod"), []byte("module github.com/pelletier/go-toml\n"), 0644)
	namesGo := extractor.DetectPackageNames(tmpGo)
	if len(namesGo) != 1 || namesGo[0] != "github.com/pelletier/go-toml" {
		t.Errorf("unexpected package names from go.mod: %v", namesGo)
	}

	// 3. Cargo.toml
	tmpRust := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpRust, "Cargo.toml"), []byte("[package]\nname = \"serde\"\nversion = \"1.0.0\"\n"), 0644)
	namesRust := extractor.DetectPackageNames(tmpRust)
	if len(namesRust) != 1 || namesRust[0] != "serde" {
		t.Errorf("unexpected package names from Cargo.toml: %v", namesRust)
	}

	// 4. pyproject.toml
	tmpPy := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpPy, "pyproject.toml"), []byte("[project]\nname = \"safety\"\nversion = \"2.0.0\"\n"), 0644)
	namesPy := extractor.DetectPackageNames(tmpPy)
	if len(namesPy) != 1 || namesPy[0] != "safety" {
		t.Errorf("unexpected package names from pyproject.toml: %v", namesPy)
	}
}

func TestIsConsumerCommand(t *testing.T) {
	packages := []string{"fork-ts-checker-webpack-plugin", "safety", "github.com/pelletier/go-toml"}

	tests := []struct {
		cmd          string
		wantConsumer bool
	}{
		{"npm install --save-dev fork-ts-checker-webpack-plugin", true},
		{"yarn add --dev fork-ts-checker-webpack-plugin", true},
		{"pnpm add fork-ts-checker-webpack-plugin", true},
		{"bun add -d fork-ts-checker-webpack-plugin", true},
		{"pip install safety", true},
		{"pip3 install --upgrade safety", true},
		{"go get github.com/pelletier/go-toml", true},
		{"npm install -g my-cli", true},
		{"yarn global add my-cli", true},
		{"npm install", false},
		{"npm ci", false},
		{"yarn install", false},
		{"pip install -r requirements.txt", false},
		{"pip install -e .", false},
		{"go mod download", false},
		{"go test ./...", false},
	}

	for _, tc := range tests {
		got := extractor.IsConsumerCommand(tc.cmd, packages, "")
		if got != tc.wantConsumer {
			t.Errorf("IsConsumerCommand(%q) = %v, want %v", tc.cmd, got, tc.wantConsumer)
		}
	}
}

func TestExtractFromDirectory_FiltersConsumerInstall(t *testing.T) {
	tempDir := t.TempDir()

	pkgJSON := `{"name": "fork-ts-checker-webpack-plugin"}`
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(pkgJSON), 0644)

	readme := `# fork-ts-checker-webpack-plugin

## Installation

` + "```bash\n" +
		"npm install --save-dev fork-ts-checker-webpack-plugin\n" +
		"yarn add --dev fork-ts-checker-webpack-plugin\n" +
		"```\n"

	contributing := `# Contributing

## Development

` + "```bash\n" +
		"yarn install\n" +
		"yarn test\n" +
		"```\n"

	_ = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte(readme), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "CONTRIBUTING.md"), []byte(contributing), 0644)

	cmds, err := extractor.ExtractFromDirectory(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should filter out the 2 consumer installation commands and only keep the 2 contributing commands
	if len(cmds) != 2 {
		t.Fatalf("expected 2 contributor commands, got %d: %+v", len(cmds), cmds)
	}
	if cmds[0].Command != "yarn install" || cmds[0].Type != extractor.StepTypeSetup {
		t.Errorf("expected yarn install, got %+v", cmds[0])
	}
	if cmds[1].Command != "yarn test" || cmds[1].Type != extractor.StepTypeTest {
		t.Errorf("expected yarn test, got %+v", cmds[1])
	}
}
