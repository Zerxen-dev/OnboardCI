# Contributing to OnboardCI

Thank you for your interest in improving OnboardCI! This document provides guidelines for local development, adding language presets, running tests, and proposing changes.

---

## Code of Conduct

All contributors and maintainers are expected to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

---

## Development Setup

### Prerequisites

- **Go**: 1.23 or newer
- **Docker**: For running containerized execution tests
- **Make**: For running project automation tasks
- **Git**

### Clone & Build

```bash
git clone https://github.com/Zerxen-dev/onboardci.git
cd onboardci
make build
```

The compiled binary will be placed at `./bin/onboardci`.

---

## Running Tests

OnboardCI utilizes Go unit tests and integration tests.

### All Unit Tests

```bash
make test-unit
```

### Full Test Suite (Unit + Integration + Race Detector)

```bash
make test
```

### Validating Brand SVGs

```bash
make assets
```

This verifies that all SVGs in `assets/` are valid XML, contain no `<script>` tags, and have zero external `http` references.

---

## How to Add a Language Preset / Command Heuristic

OnboardCI identifies setup and test commands inside fenced code blocks (`bash`, `sh`, `shell`) through heuristic matching.

To add support for a new runtime or build system (e.g. Zig, Elixir, Swift, Deno):

1. Open `internal/extractor/heuristic.go`.
2. Add relevant setup patterns (e.g. `mix deps.get`, `zig build`, `deno install`) to `setupKeywords`.
3. Add relevant test patterns (e.g. `mix test`, `zig build test`, `deno test`) to `testKeywords`.
4. Add corresponding test cases in `internal/extractor/extractor_test.go` asserting:
   - Command string is extracted accurately.
   - `StepType` is assigned correctly (`setup` vs `test`).
   - Line numbers match the source document.
5. Run `make test` to verify your changes pass.

---

## Proposing Brand & Visual Changes

OnboardCI's visual identity is strictly geometric and self-contained:
- Follow specifications in [assets/brand.md](assets/brand.md).
- Any updated SVG must be hand-authored XML without raster images, external hrefs, or embedded scripts.
- Validate via `make assets`.

---

## Setting the Repository Social Preview

GitHub does not provide a REST API endpoint for uploading repository OpenGraph social cards. Maintainers can set the card manually via the GitHub web UI:

1. Navigate to repository **Settings** (`https://github.com/Zerxen-dev/OnboardCI/settings`).
2. Under **General**, scroll down to the **Social preview** section.
3. Click **Edit** → **Upload an image...**.
4. Select `assets/social-preview.png` (1280×640 px).


---

## Submitting a Pull Request

1. Fork the repository and create a feature branch (`git checkout -b feat/my-improvement`).
2. Implement your changes following TDD (write failing tests first).
3. Ensure `make check` passes (`make test`, `make lint`, `make assets`).
4. Commit your changes with clear, descriptive commit messages.
5. Push to your fork and submit a Pull Request against `main`.
