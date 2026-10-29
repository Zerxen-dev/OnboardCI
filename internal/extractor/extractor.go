package extractor

import (
	"bufio"
	"bytes"
	"encoding/json"
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
	return ExtractFromMarkdownFiltered(filePath, content, nil)
}

// ExtractFromMarkdownFiltered scans Markdown content, filters out consumer installation
// commands matching the repository's own package name, and returns contributor commands.
func ExtractFromMarkdownFiltered(filePath string, content []byte, packageNames []string) ([]ExtractedCommand, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var commands []ExtractedCommand

	inCodeBlock := false
	currentTag := ""
	currentHeading := ""
	currentLine := 0

	for scanner.Scan() {
		currentLine++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if !inCodeBlock && strings.HasPrefix(trimmed, "#") {
			currentHeading = strings.ToLower(strings.TrimLeft(trimmed, "# "))
			continue
		}

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
				// Filter out library consumer install commands (e.g. npm install <own-pkg>)
				if IsConsumerCommand(cleaned, packageNames, currentHeading) {
					continue
				}

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

// DetectPackageNames inspects cwd for package manifests and extracts the package/module names.
func DetectPackageNames(cwd string) []string {
	var names []string

	// 1. package.json
	pkgJSONPath := filepath.Join(cwd, "package.json")
	if data, err := os.ReadFile(pkgJSONPath); err == nil {
		var p struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &p); err == nil && p.Name != "" {
			names = append(names, p.Name)
			// If scoped package (e.g. @org/foo), also add bare name (foo)
			if strings.HasPrefix(p.Name, "@") && strings.Contains(p.Name, "/") {
				parts := strings.SplitN(p.Name, "/", 2)
				if len(parts) == 2 && parts[1] != "" {
					names = append(names, parts[1])
				}
			}
		}
	}

	// 2. go.mod
	goModPath := filepath.Join(cwd, "go.mod")
	if data, err := os.ReadFile(goModPath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "module ") {
				modName := strings.TrimSpace(strings.TrimPrefix(line, "module "))
				if modName != "" {
					names = append(names, modName)
				}
				break
			}
		}
	}

	// 3. Cargo.toml
	cargoPath := filepath.Join(cwd, "Cargo.toml")
	if data, err := os.ReadFile(cargoPath); err == nil {
		var inPackage bool
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "[") {
				inPackage = (line == "[package]")
				continue
			}
			if inPackage && strings.HasPrefix(line, "name") && strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
				if val != "" {
					names = append(names, val)
				}
				break
			}
		}
	}

	// 4. pyproject.toml / setup.py / setup.cfg
	pyprojectPath := filepath.Join(cwd, "pyproject.toml")
	if data, err := os.ReadFile(pyprojectPath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "name") && strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
				if val != "" {
					names = append(names, val)
					break
				}
			}
		}
	}

	return names
}

// IsConsumerCommand checks if a command is intended for library consumers rather than repo contributors.
func IsConsumerCommand(cmd string, packageNames []string, sectionHeading string) bool {
	lower := strings.ToLower(cmd)

	// Global install commands are for users/consumers, not contributors
	if strings.Contains(lower, "npm install -g") || strings.Contains(lower, "npm i -g") ||
		strings.Contains(lower, "yarn global add") || strings.Contains(lower, "pnpm add -g") ||
		strings.Contains(lower, "bun add -g") {
		return true
	}

	// Check if command is installing the repository's own package name
	for _, pkg := range packageNames {
		lowerPkg := strings.ToLower(pkg)
		if !strings.Contains(lower, lowerPkg) {
			continue
		}

		// Ensure it's being installed/added as a dependency argument
		installPatterns := []string{
			"npm install", "npm i",
			"yarn add",
			"pnpm add", "pnpm install", "pnpm i",
			"bun add",
			"pip install", "pip3 install", "python -m pip install", "python3 -m pip install",
			"go get",
			"cargo add",
		}

		for _, pattern := range installPatterns {
			if strings.Contains(lower, pattern) {
				// Don't flag local editable installs like "pip install -e ." or "npm install ."
				if strings.Contains(lower, "-e .") || strings.Contains(lower, " . ") || strings.HasSuffix(lower, " .") {
					continue
				}
				// Verify the package name appears after the install/add pattern
				idx := strings.Index(lower, pattern)
				after := lower[idx+len(pattern):]
				if strings.Contains(after, lowerPkg) {
					return true
				}
			}
		}
	}

	return false
}

// ExtractFromDirectory inspects README.md, CONTRIBUTING.md, and docs in cwd.
func ExtractFromDirectory(cwd string) ([]ExtractedCommand, error) {
	packageNames := DetectPackageNames(cwd)

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

		cmds, err := ExtractFromMarkdownFiltered(fullPath, data, packageNames)
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
		} else if pkgJSONPath := filepath.Join(cwd, "package.json"); fileExists(pkgJSONPath) {
			setupCmd := "npm install"
			if fileExists(filepath.Join(cwd, "package-lock.json")) {
				setupCmd = "npm ci"
			} else if fileExists(filepath.Join(cwd, "yarn.lock")) {
				setupCmd = "yarn install"
			} else if fileExists(filepath.Join(cwd, "pnpm-lock.yaml")) {
				setupCmd = "pnpm install"
			}
			allCommands = append(allCommands,
				ExtractedCommand{
					SourceFile: "package.json",
					LineNumber: 1,
					Type:       StepTypeSetup,
					Command:    setupCmd,
				},
				ExtractedCommand{
					SourceFile: "package.json",
					LineNumber: 1,
					Type:       StepTypeTest,
					Command:    "npm test",
				},
			)
		} else if goModPath := filepath.Join(cwd, "go.mod"); fileExists(goModPath) {
			allCommands = append(allCommands,
				ExtractedCommand{
					SourceFile: "go.mod",
					LineNumber: 1,
					Type:       StepTypeSetup,
					Command:    "go mod download",
				},
				ExtractedCommand{
					SourceFile: "go.mod",
					LineNumber: 1,
					Type:       StepTypeTest,
					Command:    "go test ./...",
				},
			)
		} else if cargoPath := filepath.Join(cwd, "Cargo.toml"); fileExists(cargoPath) {
			allCommands = append(allCommands,
				ExtractedCommand{
					SourceFile: "Cargo.toml",
					LineNumber: 1,
					Type:       StepTypeSetup,
					Command:    "cargo build",
				},
				ExtractedCommand{
					SourceFile: "Cargo.toml",
					LineNumber: 1,
					Type:       StepTypeTest,
					Command:    "cargo test",
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
