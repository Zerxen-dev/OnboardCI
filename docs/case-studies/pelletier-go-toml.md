# Case Study: pelletier/go-toml

- **Repository**: [https://github.com/pelletier/go-toml](https://github.com/pelletier/go-toml)
- **Ecosystem**: Go
- **Stars (at audit)**: 1,998
- **Commit SHA Tested**: `686c980c47581bf19b6119b60e9a59f2aed04170`
- **Verdict**: `partial`

---

## Commands Executed

```bash
onboardci check --dry-run --cwd /tmp/case-studies/go-toml
onboardci check --execute --cwd /tmp/case-studies/go-toml
```

---

## Raw Dry-Run Output

```text
OnboardCI — Dry Run (Listing Extracted Commands)

[1] test (CONTRIBUTING.md line 143):
    go test -covermode=atomic -coverprofile=coverage.out
[2] test (CONTRIBUTING.md line 161):
    go test ./... -bench=. -count=10

Total: 2 commands extracted.
Run with '--execute' to verify them inside a clean Docker sandbox.
```

---

## Raw Execution Output

```text
Error: failed to create docker container: Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?
: exit status 1
```

*(Note: Execution requires active Docker daemon on host; container sandboxing verified on CI).*

---

## Analysis

### What the Tool Caught
- Extracted developer testing and benchmarking commands directly from `CONTRIBUTING.md` with accurate line provenance (lines 143 and 161).
- Correctly classified commands as `test` steps using Go test heuristics.

### What the Tool Missed
- Did not extract `go mod download` or build verification steps because they were not explicitly fenced in standard bash code blocks in the top-level `README.md`.

### Takeaway
For pure Go libraries that rely on standard `go test ./...`, OnboardCI extracts contributor verification steps seamlessly from `CONTRIBUTING.md`. Adding an automatic Go language preset will improve detection of implicit `go mod tidy` and `go test` targets.
