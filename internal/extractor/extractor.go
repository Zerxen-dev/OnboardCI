package extractor

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StepType denotes whether a command is setup or verification/test.
type StepType string

const (
	StepTypeSetup StepType = "setup"
	StepTypeTest  StepType = "test"
)

// ExtractedCommand encapsulates an extracted shell command with source provenance.
type ExtractedCommand struct {
	SourceFile string   `json:"source_file"`
	LineNumber int      `json:"line_number"`
	Type       StepType `json:"type"`
	Command    string   `json:"command"`
}

var supportedCodeFenceTags = map[string]bool{
	"bash":    true,
	"sh":      true,
	"shell":   true,
	"zsh":     true,
	"console": true,
	"":        true, // untagged blocks may contain recognized shell commands
}

// ExtractFromMarkdown scans Markdown content and returns commands with exact line references.
func ExtractFromMarkdown(filePath string, content []byte) ([]ExtractedCommand, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var commands []ExtractedCommand

	inCodeBlock := false
	currentTag := ""
	currentLine := 0

	for scanner.Scan() {
		currentLine++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				// closing block
				inCodeBlock = false
				currentTag = ""
			} else {
				// opening block
				inCodeBlock = true
				rawTag := strings.TrimPrefix(trimmed, "```")
				currentTag = strings.ToLower(strings.TrimSpace(rawTag))
			}
			continue
		}

		if inCodeBlock {
			if !supportedCodeFenceTags[currentTag] {
				continue
			}

			cleaned := CleanCommandLine(trimmed)
			if cleaned == "" || IsComment(cleaned) {
				continue
			}

			stepType, isRecognized := ClassifyCommand(cleaned)
			if isRecognized {
				commands = append(commands, ExtractedCommand{
					SourceFile: filePath,
					LineNumber: currentLine,
					Type:       stepType,
					Command:    cleaned,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filePath, err)
	}

	return commands, nil
}

// ExtractFromDirectory inspects README.md, CONTRIBUTING.md, and docs in cwd.
func ExtractFromDirectory(cwd string) ([]ExtractedCommand, error) {
	docFiles := []string{
		"README.md",
		"readme.md",
		"CONTRIBUTING.md",
		"contributing.md",
	}

	var allCommands []ExtractedCommand
	var seenStats []os.FileInfo

	for _, rel := range docFiles {
		fullPath := filepath.Join(cwd, rel)
		stat, err := os.Stat(fullPath)
		if err != nil || stat.IsDir() {
			continue
		}

		// Check if we already processed this physical file (handles case-insensitive filesystems)
		alreadySeen := false
		for _, seen := range seenStats {
			if os.SameFile(seen, stat) {
				alreadySeen = true
				break
			}
		}
		if alreadySeen {
			continue
		}
		seenStats = append(seenStats, stat)

		data, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read doc file %s: %w", fullPath, err)
		}

		cmds, err := ExtractFromMarkdown(fullPath, data)
		if err != nil {
			return nil, err
		}

		allCommands = append(allCommands, cmds...)
	}

	// Fallback to manifest-based extraction when docs do not contain executable commands
	if len(allCommands) == 0 {
		reqPath := filepath.Join(cwd, "requirements.txt")
		if fileExists(reqPath) {
			allCommands = append(allCommands,
				ExtractedCommand{
					SourceFile: "requirements.txt",
					LineNumber: 1,
					Type:       StepTypeSetup,
					Command:    "pip install -r requirements.txt",
				},
				ExtractedCommand{
					SourceFile: "requirements.txt",
					LineNumber: 1,
					Type:       StepTypeTest,
					Command:    "pytest",
				},
			)
		} else if pyprojPath := filepath.Join(cwd, "pyproject.toml"); fileExists(pyprojPath) {
			setupCmd := "pip install ."
			if fileExists(filepath.Join(cwd, "poetry.lock")) {
				setupCmd = "poetry install"
			}
			allCommands = append(allCommands,
				ExtractedCommand{
					SourceFile: "pyproject.toml",
					LineNumber: 1,
					Type:       StepTypeSetup,
					Command:    setupCmd,
				},
				ExtractedCommand{
					SourceFile: "pyproject.toml",
					LineNumber: 1,
					Type:       StepTypeTest,
					Command:    "pytest",
				},
			)
		}
	}

	return allCommands, nil
}

func fileExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}
