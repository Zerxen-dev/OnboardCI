package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	envRegexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)missing env(?:ironment)? var(?:iable)?[:\s]+['"]?([A-Za-z0-9_]+)['"]?`),
		regexp.MustCompile(`(?i)(?:env(?:ironment)? variable )?['"]?([A-Za-z0-9_]+)['"]? is not set`),
		regexp.MustCompile(`(?i)(?:env(?:ironment)? variable )?['"]?([A-Za-z0-9_]+)['"]? is not defined`),
		regexp.MustCompile(`(?i)KeyError:\s*['"]([A-Za-z0-9_]+)['"]`),
		regexp.MustCompile(`(?i)required environment variable ['"]([A-Za-z0-9_]+)['"]`),
		regexp.MustCompile(`(?i)process\.env\.([A-Za-z0-9_]+)`),
	}

	depRegexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)Cannot find module ['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)ModuleNotFoundError: No module named ['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)ImportError: No module named ['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)command not found:\s*(\S+)`),
		regexp.MustCompile(`(?i)/bin/sh:\s*(\S+):\s*not found`),
	}

	fileRegexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)ENOENT: no such file or directory[,\s]+(?:open )?['"]?([^'"\n]+)['"]?`),
		regexp.MustCompile(`(?i)No such file or directory:\s*['"]?([^'"\n]+)['"]?`),
	}

	connRegexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)ECONNREFUSED`),
		regexp.MustCompile(`(?i)Connection refused`),
		regexp.MustCompile(`(?i)dial tcp.*connection refused`),
	}
)

// findEnvTemplate returns the path to .env.example or .env.sample in the current
// working directory, or "" if neither exists.
func findEnvTemplate() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	for _, name := range []string{".env.example", ".env.sample"} {
		path := filepath.Join(cwd, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return ""
}

// envTemplateHasVar reports whether the given variable name is documented in the
// template file. Caps file reading at 1MB to prevent unbounded memory consumption.
func envTemplateHasVar(path, varName string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() > 1024*1024 {
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip optional "export " prefix.
		line = strings.TrimPrefix(line, "export ")
		if idx := strings.Index(line, "="); idx >= 0 {
			key := strings.TrimSpace(line[:idx])
			if key == varName {
				return true
			}
		}
	}
	return false
}

// DiagnoseFailure analyzes output text and returns a human-friendly failure reason and suggestion.
func DiagnoseFailure(output string, exitCode int) (string, string) {
	combined := output

	// Check missing environment variables
	for _, re := range envRegexes {
		matches := re.FindStringSubmatch(combined)
		if len(matches) > 1 {
			varName := strings.TrimSpace(matches[1])
			suggestion := "document required env vars or provide .env.example"
			if tmpl := findEnvTemplate(); tmpl != "" && envTemplateHasVar(tmpl, varName) {
				base := filepath.Base(tmpl)
				suggestion = fmt.Sprintf("Found %s with missing variable. Run cp %s .env and document required values.", base, base)
			}
			return fmt.Sprintf("missing env var %s", varName), suggestion
		}
	}

	// Check missing dependencies
	for _, re := range depRegexes {
		matches := re.FindStringSubmatch(combined)
		if len(matches) > 1 {
			pkg := strings.TrimSpace(matches[1])
			return fmt.Sprintf("missing dependency or tool '%s'", pkg), "document installation prerequisites or add to package manifest"
		}
	}

	// Check missing files
	for _, re := range fileRegexes {
		matches := re.FindStringSubmatch(combined)
		if len(matches) > 1 {
			filePath := strings.TrimSpace(matches[1])
			return fmt.Sprintf("missing file or directory '%s'", filePath), "commit the required template or document file creation in setup"
		}
	}

	// Check connection failures
	for _, re := range connRegexes {
		if re.MatchString(combined) {
			return "connection refused (external service or daemon not accessible)", "mock external service in tests or document local service startup"
		}
	}

	// Fallback
	return fmt.Sprintf("command exited with status code %d", exitCode), "review error logs above for missing prerequisites or test assertions"
}
