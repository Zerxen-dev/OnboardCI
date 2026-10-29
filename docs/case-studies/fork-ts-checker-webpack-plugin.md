# Case Study: TypeStrong/fork-ts-checker-webpack-plugin

- **Repository**: [https://github.com/TypeStrong/fork-ts-checker-webpack-plugin](https://github.com/TypeStrong/fork-ts-checker-webpack-plugin)
- **Ecosystem**: JavaScript / TypeScript
- **Stars (at audit)**: 2,000
- **Commit SHA Tested**: `9f70a3dcdaf216177b9b7b85426fc8e473bfad4e`
- **Verdict**: `resolved` (Fixed in [Issue #6](https://github.com/Zerxen-dev/OnboardCI/issues/6))

---

## Commands Executed

```bash
onboardci check --dry-run --cwd /tmp/case-studies/fork-ts-checker
onboardci check --execute --cwd /tmp/case-studies/fork-ts-checker
```

---

## Raw Dry-Run Output

```text
OnboardCI — Dry Run (Listing Extracted Commands)

[1] setup (README.md line 35):
    npm install --save-dev fork-ts-checker-webpack-plugin
[2] setup (README.md line 38):
    yarn add --dev fork-ts-checker-webpack-plugin
[3] setup (CONTRIBUTING.md line 62):
    yarn install
[4] setup (CONTRIBUTING.md line 197):
    yarn install

Total: 4 commands extracted.
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

## Analysis & Bug Identification

### What the Tool Caught
- Extracted `yarn install` from `CONTRIBUTING.md` lines 62 and 197.

### What the Tool Got Wrong (Bug Identified)
- In `README.md`, lines 35 and 38 are **consumer installation instructions** for users installing the plugin into their own webpack project:
  ```bash
  npm install --save-dev fork-ts-checker-webpack-plugin
  yarn add --dev fork-ts-checker-webpack-plugin
  ```
- OnboardCI extracted both as **contributor setup commands** because they contain `npm install` and `yarn add`.
- Executing `npm install <self-package>` in a developer sandbox is circular and incorrect.

### Resolution
- Fixed in [Issue #6](https://github.com/Zerxen-dev/OnboardCI/issues/6): `ExtractFromDirectory` automatically detects `package.json` package names and filters out self-referencing consumer installation commands (`npm install <own-pkg>`, `yarn add <own-pkg>`, etc.), accurately extracting only genuine contributor setup commands (`yarn install`).
