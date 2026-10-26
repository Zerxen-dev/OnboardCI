# Case Study: pyupio/safety

- **Repository**: [https://github.com/pyupio/safety](https://github.com/pyupio/safety)
- **Ecosystem**: Python
- **Stars (at audit)**: 2,000
- **Commit SHA Tested**: `997602736a20c6f4d9fec8c2a5d479624b9ddb9a`
- **Verdict**: `partial`

---

## Commands Executed

```bash
onboardci check --dry-run --cwd /tmp/case-studies/safety
onboardci check --execute --cwd /tmp/case-studies/safety
```

---

## Raw Dry-Run Output

```text
OnboardCI — Dry Run (Listing Extracted Commands)

[1] setup (CONTRIBUTING.md line 81):
    pip install -r requirements.txt

Total: 1 commands extracted.
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
- Extracted contributor environment setup (`pip install -r requirements.txt`) from line 81 of `CONTRIBUTING.md`.
- Correctly classified command as `setup`.

### What the Tool Missed
- Did not extract test commands (`pytest` or `tox`) because the repository mentions test commands inside inline code spans and documentation tables rather than isolated multi-line fenced code blocks.
- OnboardCI currently defaults its Docker sandbox image to Node.js; running Python projects under `--execute` requires the Python language preset tracked in [Issue #1](https://github.com/Zerxen-dev/OnboardCI/issues/1).

### Takeaway
Validates the urgent need for [Issue #1 (Python/pip language preset)](https://github.com/Zerxen-dev/OnboardCI/issues/1) so that Python virtualenvs, `requirements.txt`, and `pytest` are automatically executed inside a Python 3.12 container sandbox.
