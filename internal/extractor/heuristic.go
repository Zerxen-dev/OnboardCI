package extractor

import (
	"strings"
)

// CleanCommandLine removes prompts (like "$ " or "> ") and trims whitespace.
func CleanCommandLine(raw string) string {
	line := strings.TrimSpace(raw)
	if strings.HasPrefix(line, "$ ") {
		line = strings.TrimSpace(line[2:])
	} else if strings.HasPrefix(line, "$") {
		line = strings.TrimSpace(line[1:])
	} else if strings.HasPrefix(line, "> ") {
		line = strings.TrimSpace(line[2:])
	}
	return line
}

// IsComment checks if a line is a shell comment.
func IsComment(line string) bool {
	return strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//")
}

// ClassifyCommand inspects command tokens and identifies setup vs test commands.
func ClassifyCommand(cmd string) (StepType, bool) {
	lower := strings.ToLower(cmd)

	// Test patterns take precedence when explicit test keywords exist
	testKeywords := []string{
		"npm test", "npm run test", "npm t",
		"yarn test", "pnpm test",
		"pytest", "python -m unittest", "python -m pytest",
		"python3 -m unittest", "python3 -m pytest", "py.test",
		"poetry run pytest", "pipenv run pytest", "tox", "nox",
		"python setup.py test", "python3 setup.py test",
		"cargo test",
		"go test",
		"make test", "make check",
		"ctest",
		"rake test", "bundle exec rspec",
		"mvn test", "gradle test",
		"dotnet test",
	}

	for _, kw := range testKeywords {
		if strings.Contains(lower, kw) {
			return StepTypeTest, true
		}
	}

	// Setup patterns
	setupKeywords := []string{
		"npm ci", "npm install", "npm i", "npm run build",
		"yarn install", "yarn build", "yarn add",
		"pnpm install", "pnpm i", "pnpm build",
		"pip install", "pip3 install", "python -m pip install", "python3 -m pip install",
		"poetry install", "pipenv install", "pip-sync", "uv pip install", "uv sync",
		"python setup.py install", "python setup.py develop", "flit install",
		"bundle install",
		"cargo build",
		"go mod download", "go get", "go build",
		"make setup", "make build", "make install", "make deps", "make",
		"mvn clean install", "gradle build",
		"dotnet restore", "dotnet build",
		"composer install",
	}

	for _, kw := range setupKeywords {
		if strings.Contains(lower, kw) {
			return StepTypeSetup, true
		}
	}

	return "", false
}
