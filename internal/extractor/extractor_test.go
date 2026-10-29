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
