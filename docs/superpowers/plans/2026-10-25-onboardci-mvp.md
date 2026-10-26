# OnboardCI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build, test, brand, and publish `OnboardCI`, a production-ready, zero-telemetry open-source CLI tool and GitHub Action that verifies if a new contributor can set up an open-source repository by following only documented setup and test steps inside a clean Docker container.

**Architecture:** A single-binary Go CLI using `cobra` for command routing, a resilient Markdown AST/line-tracking parser for extracting fenced setup/test commands with exact line references, a YAML config loader for `.onboardci.yml`, a sandboxed Docker runner with context timeouts, memory/CPU bounds, and non-root execution, and pluggable formatters (human-readable CLI diagnostic with actionable suggestions, JSON for automation, and GitHub PR comments). Hand-authored SVG visual identity and complete OSS community scaffolding.

**Tech Stack:** Go (1.23+), `spf13/cobra`, `gopkg.in/yaml.v3`, Docker CLI / Engine API, GitHub Actions, Hand-authored SVG.

**Spec:** `/storage/emulated/0/Onboardci/prompt.txt`

## Global Constraints
- Single Go binary; no runtime dependencies beyond Docker.
- Dry-run is default (`onboardci check`); `--execute` must be explicitly specified to run Docker.
- No telemetry, no user accounts, no remote calls beyond Docker image pulls.
- Apache-2.0 License.
- Non-root user execution inside container (`--user 1000:1000`).
- Context cancellation enforced timeouts (default `10m`).
- CPU and memory resource limits enforced on container.
- No host bind mounts: target repository is copied into container workspace.
- Original geometric logo mark on 512x512 grid; strictly hand-authored clean SVGs without external links or scripts.
- Codebase must be clean, modular, and maintainable by a weekend open-source contributor.

## Workspace Inspection Findings
- Working Directory: `/storage/emulated/0/Onboardci`
- Current Workspace Files: Only `prompt.txt` is present.
- Environment: Ubuntu 26.04 LTS (arm64) on Linux kernel.
- Installed Toolchain: `git` (v2.53.0), `gh` (v2.46.0, authenticated to GitHub user `Zerxen-dev` with `repo` scopes), `make`.
- Dependencies to provision for build: `golang-go` package (available via Ubuntu resolute repo). Docker daemon/mock executor abstraction for headless/arm64 environments.

---

## Directory Structure

```text
/storage/emulated/0/Onboardci/
├── .github/
│   └── workflows/
│       ├── test.yml                 # CI unit & integration test runner
│       └── release.yml              # Release binary build & GitHub release
├── action.yml                       # GitHub Action wrapper for PR checks
├── cmd/
│   └── onboardci/
│       └── main.go                  # CLI entrypoint
├── internal/
│   ├── cli/
│   │   ├── root.go                  # Root cobra command, global flags (--cwd, --config, --timeout)
│   │   ├── check.go                 # 'check' command (--execute, --json)
│   │   └── version.go               # 'version' command
│   ├── config/
│   │   ├── config.go                # .onboardci.yml struct schema, defaults, loader
│   │   └── config_test.go           # Unit tests for config parsing & fallback
│   ├── extractor/
│   │   ├── extractor.go             # Markdown parser extracting fenced commands with line numbers
│   │   ├── heuristic.go             # Setup vs test classification heuristics
│   │   └── extractor_test.go        # Unit tests on README/CONTRIBUTING fixtures
│   ├── runner/
│   │   ├── runner.go                # Runner interface & execution model
│   │   ├── docker.go                # Docker CLI runner (non-root, timeout, copy repo, limits)
│   │   ├── mock.go                  # Mock runner for unit testing & environments without Docker daemon
│   │   └── docker_test.go           # Runner unit tests
│   └── reporter/
│       ├── reporter.go              # Reporter interface & diagnosis suggestions generator
│       ├── text.go                  # Plain-text diagnostic output (line, command, reason, suggestion)
│       ├── json.go                  # Structured JSON output format
│       ├── github.go                # GitHub PR comment markdown generator
│       └── reporter_test.go         # Reporter output formatting tests
├── assets/
│   ├── logo.svg                     # Primary geometric checkpoint mark (512x512)
│   ├── logo-wordmark.svg            # Mark + "OnboardCI" typography
│   ├── logo-mono.svg                # Single-color contrast version for dark/light themes
│   ├── favicon.svg                  # 32x32 simplified mark
│   ├── social-preview.svg           # 1280x640 GitHub social preview banner
│   ├── brand.md                     # Brand guidelines, palette hex codes, clear space, dos/don'ts
│   └── logo-prompt.txt              # Future variant generation prompt documentation
├── testdata/
│   ├── fixtures/
│   │   ├── broken-repo/             # Fixture repo with missing env var in README
│   │   │   ├── README.md
│   │   │   └── package.json
│   │   ├── valid-repo/              # Fixture repo with valid .onboardci.yml
│   │   │   ├── README.md
│   │   │   └── .onboardci.yml
│   │   └── sample-docs/
│   │       ├── README.md
│   │       └── CONTRIBUTING.md
├── docs/
│   └── non-goals.md                 # Scope boundaries: not a security scanner, not a CI replacement
├── LICENSE                          # Apache License 2.0
├── README.md                        # Project landing page with centered logo, install, usage, config
├── CONTRIBUTING.md                  # Contributor workflow, adding language presets, testing guide
├── SECURITY.md                      # Security policy & responsible disclosure
├── CODE_OF_CONDUCT.md               # Contributor Covenant v2.1
├── Makefile                         # make build, test, lint, install, assets, check
├── .gitignore                       # Binaries, temp test outputs, OS artifacts
├── go.mod                           # Go module definition (github.com/Zerxen-dev/onboardci)
└── go.sum
```

---

## Detailed Task Breakdown

### Phase 2: Build the Core MVP

#### Task 1: Scaffolding, Go Module & Config Schema
**Files:**
- Create: `go.mod`, `go.sum`, `internal/config/config.go`, `internal/config/config_test.go`
**Interfaces:**
- `Config` struct:
  ```go
  type Config struct {
      Setup  []string     `yaml:"setup"`
      Test   []string     `yaml:"test"`
      Env    []string     `yaml:"env"`
      Docker DockerConfig `yaml:"docker"`
  }
  type DockerConfig struct {
      Image   string        `yaml:"image"`
      Timeout time.Duration `yaml:"timeout"`
      User    string        `yaml:"user"`
      CPUs    string        `yaml:"cpus"`
      Memory  string        `yaml:"memory"`
  }
  func LoadConfig(cwd string, customPath string) (*Config, string, error)
  ```
- [ ] **Step 1: Write failing config tests** (test default fallbacks, valid `.onboardci.yml`, and syntax errors).
- [ ] **Step 2: Run `go test ./internal/config` to verify failure.**
- [ ] **Step 3: Implement `LoadConfig` with YAML unmarshaling and sensible defaults (`node:20-alpine`, `10m` timeout, `1000:1000` user).**
- [ ] **Step 4: Run tests and verify PASS.**

#### Task 2: Markdown Command Extractor with Source Line Tracking
**Files:**
- Create: `internal/extractor/extractor.go`, `internal/extractor/heuristic.go`, `internal/extractor/extractor_test.go`
**Interfaces:**
- `CommandStep` struct:
  ```go
  type StepType string // "setup" | "test"
  type ExtractedCommand struct {
      SourceFile string
      LineNumber int
      Type       StepType
      Command    string
  }
  func ExtractFromMarkdown(filePath string, content []byte) ([]ExtractedCommand, error)
  func ExtractFromDirectory(cwd string) ([]ExtractedCommand, error)
  ```
- Heuristics:
  - Fenced code blocks with tags: `bash`, `sh`, `shell`, `zsh`, `console`, or untagged if matching shell keywords.
  - Setup keywords: `npm install`, `npm ci`, `yarn install`, `pnpm install`, `pip install`, `poetry install`, `bundle install`, `cargo build`, `go mod download`, `make setup`, `make build`, etc.
  - Test keywords: `npm test`, `yarn test`, `pnpm test`, `pytest`, `cargo test`, `go test`, `make test`, `ctest`, etc.
  - Line numbers tracked accurately from opening fence to individual shell lines.
- [ ] **Step 1: Write failing extractor unit tests with multi-block Markdown fixtures.**
- [ ] **Step 2: Verify test fails.**
- [ ] **Step 3: Implement line-indexed Markdown scanner.**
- [ ] **Step 4: Verify test passes.**

#### Task 3: Docker & Isolation Execution Engine
**Files:**
- Create: `internal/runner/runner.go`, `internal/runner/docker.go`, `internal/runner/mock.go`, `internal/runner/docker_test.go`
**Interfaces:**
- `Runner` interface:
  ```go
  type CommandResult struct {
      Command        string
      StepType       string
      SourceFile     string
      LineNumber     int
      ExitCode       int
      Stdout         string
      Stderr         string
      Duration       time.Duration
      FailureReason  string
      Suggestion     string
  }
  type Runner interface {
      Execute(ctx context.Context, cwd string, cfg *config.Config, cmds []extractor.ExtractedCommand) ([]CommandResult, error)
  }
  ```
- Runner implementation details:
  - Validates Docker availability.
  - Spins up container with `--user 1000:1000`, `--cpus 2.0`, `--memory 2g`, `--rm`, `--network=bridge`.
  - Copies repository into isolated container directory `/workspace` (NO host bind mounts).
  - Executes setup commands sequentially, then test commands.
  - Context cancellation enforces `--timeout` (default 10m).
  - Truncates logs at 64KB per step.
  - Intelligently analyzes stderr for common onboarding blockers (missing env var, missing dependency, unbound port, missing file) and generates suggestions.
- [ ] **Step 1: Write failing runner unit tests.**
- [ ] **Step 2: Implement Docker execution engine and failure diagnosis parser.**
- [ ] **Step 3: Run unit tests and confirm pass.**

#### Task 4: CLI Surface & Report Formatters
**Files:**
- Create: `internal/cli/root.go`, `internal/cli/check.go`, `internal/cli/version.go`, `cmd/onboardci/main.go`, `internal/reporter/text.go`, `internal/reporter/json.go`, `internal/reporter/github.go`
**Commands & Flags:**
- `onboardci check`:
  - `--execute` (bool, default `false` -> dry-run mode)
  - `--json` (bool, default `false`)
  - `--cwd <path>` (string, default `.`)
  - `--timeout <duration>` (duration, default `10m`)
  - `--config <path>` (string, default empty -> auto-detect `.onboardci.yml`)
- Text Reporter:
  ```text
  FAIL: README.md line 42
  Command: npm test
  Reason: missing env var DATABASE_URL
  Suggestion: document required env vars or provide .env.example
  ```
- JSON Reporter: structured machine-readable payload.
- GitHub PR Reporter: Markdown summary table with status icons, doc lines, and foldable logs.
- [ ] **Step 1: Write unit tests for text, JSON, and GitHub reporters.**
- [ ] **Step 2: Implement CLI commands using Cobra.**
- [ ] **Step 3: Verify dry-run outputs extracted commands without container execution.**
- [ ] **Step 4: Verify `--json` and plain text reporter outputs.**

---

### Phase 3: Integration & Comprehensive Testing

#### Task 5: Integration Test Fixtures & Test Harness
**Files:**
- Create: `testdata/fixtures/broken-repo/README.md`, `testdata/fixtures/broken-repo/package.json`, `testdata/fixtures/valid-repo/README.md`, `testdata/fixtures/valid-repo/.onboardci.yml`, `internal/runner/integration_test.go`
- Target:
  - Integration test gated by `//go:build integration`.
  - Runs `onboardci check --execute` against `broken-repo` fixture.
  - Asserts non-zero exit code, error line pointing to README.md line, and correct missing env suggestion.
  - Runs `onboardci check` (dry-run) against `broken-repo` fixture and asserts exit code 0 and command extraction list.
- [ ] **Step 1: Create fixture repositories.**
- [ ] **Step 2: Write integration tests.**
- [ ] **Step 3: Wire into `Makefile` (`make test` runs standard tests + integration).**

---

### Phase 4: Brand Identity & Vector Assets

#### Task 6: Hand-Authored SVG Suite & Brand Guide
**Files:**
- Create: `assets/logo.svg`, `assets/logo-wordmark.svg`, `assets/logo-mono.svg`, `assets/favicon.svg`, `assets/social-preview.svg`, `assets/brand.md`, `assets/logo-prompt.txt`
**Design Specifications:**
- **Concept:** Checkpoint gate — doorway/threshold shape reading as a bold checkmark, framed by two vertical container wall bars. Communicates "pass through / verify at the gate".
- **Grid:** 512×512 grid, 64px safe margin. Only circles, rounded rectangles, and straight lines. No text in primary mark.
- **Palette:**
  - Deep ink `#0B1220` (Dark canvas / container)
  - Signal green `#22C55E` (Pass check / gate threshold)
  - Warning amber `#F59E0B` (Fail state)
  - Paper white `#F8FAFC` (Light background)
  - Slate `#64748B` (Structure & borders)
- **Files:**
  1. `assets/logo.svg`: 512x512 scalable primary mark.
  2. `assets/logo-wordmark.svg`: Mark + clean geometric "OnboardCI" wordmark.
  3. `assets/logo-mono.svg`: High-contrast single-color monochrome mark legible on both `#0B1220` and `#F8FAFC`.
  4. `assets/favicon.svg`: 32x32 simplified crisp mark.
  5. `assets/social-preview.svg`: 1280x640 GitHub social banner.
  6. `assets/brand.md`: Color codes, clear space, typography stack, usage rules.
  7. `assets/logo-prompt.txt`: Standalone prompt for future variant documentation.
- **Verification:**
  - XML syntax validity.
  - No `<script>`, no external links (`grep -R "http" assets/*.svg` returns 0).
  - Legible at 32px, 128px, 512px.

---

### Phase 5: GitHub Action & Open-Source Scaffolding

#### Task 7: GitHub Action & Repository Hygiene
**Files:**
- Create: `action.yml`, `.github/workflows/test.yml`, `.github/workflows/release.yml`, `README.md`, `CONTRIBUTING.md`, `LICENSE`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, `Makefile`, `.gitignore`, `docs/non-goals.md`
- `action.yml`:
  - Runs OnboardCI check in PR context and leaves a status comment.
- `README.md`:
  - Primary logo at top, centered, width 160.
  - Problem statement, quickstart, installation, `.onboardci.yml` schema, CLI flags, GitHub Action integration, and brand link.
- `Makefile`:
  - `build`, `test`, `lint`, `install`, `assets`, `clean`.
- `LICENSE`: Apache 2.0.

---

### Phase 6: Verification Loop

1. Run `make test` (`go test -race ./...`). Fix any failures.
2. Run `onboardci check --cwd testdata/fixtures/broken-repo` (dry-run). Confirm it lists commands without executing.
3. Run `onboardci check --execute --cwd testdata/fixtures/broken-repo` (execute). Confirm it fails with expected failure line.
4. Run logo verification (XML validity, contrast checks, 0 external URLs).

---

### Phase 7: GitHub Publication

1. `git init && git add -A && git commit -m "feat: initial OnboardCI MVP with brand assets"`
2. Verify `gh auth status`.
3. Create public repository: `gh repo create OnboardCI --public --source=. --remote=origin --push`.
4. Set description: "Verify that new contributors can actually set up your open-source repo by following your docs."
5. Set topics: `developer-experience`, `open-source`, `onboarding`, `ci`, `cli`, `golang`.
6. Print final summary report with repository URL and test logs.
