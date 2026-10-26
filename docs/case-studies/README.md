# Real-World Case Studies

To verify OnboardCI against production codebases outside our test fixtures, we tested the CLI against three public, popular open-source repositories spanning Go, Python, and TypeScript ecosystems (none owned by the maintainer). These studies evaluate command extraction accuracy, container sandboxing expectations, and heuristic edge cases under real documentation conventions.

---

## Case Studies Index

| Repository | Ecosystem | Stars | Verdict | Case Study | Issues Filed |
| :--- | :--- | :---: | :---: | :--- | :--- |
| **[pelletier/go-toml](https://github.com/pelletier/go-toml)** | Go | 1,998 | `partial` | [pelletier-go-toml.md](pelletier-go-toml.md) | — |
| **[pyupio/safety](https://github.com/pyupio/safety)** | Python | 2,000 | `partial` | [pyupio-safety.md](pyupio-safety.md) | [Issue #1](https://github.com/Zerxen-dev/OnboardCI/issues/1) |
| **[TypeStrong/fork-ts-checker-webpack-plugin](https://github.com/TypeStrong/fork-ts-checker-webpack-plugin)** | TypeScript | 2,000 | `bug found` | [fork-ts-checker-webpack-plugin.md](fork-ts-checker-webpack-plugin.md) | [Issue #6](https://github.com/Zerxen-dev/OnboardCI/issues/6) |

---

## Key Learnings

1. **Extraction Accuracy on Contributor Docs**: When documentation contains a dedicated `CONTRIBUTING.md`, OnboardCI reliably discovers setup and test commands (`yarn install`, `pip install -r requirements.txt`, `go test`).
2. **Library vs. Contributor Disambiguation**: In popular npm packages, `README.md` installation commands are intended for package consumers (`npm install <package>`). OnboardCI must filter self-referencing package installations to avoid circular setup failures ([Issue #6](https://github.com/Zerxen-dev/OnboardCI/issues/6)).
3. **Multi-Ecosystem Sandboxes**: Python and Go repos benefit directly from native runtime presets rather than relying on default container images ([Issue #1](https://github.com/Zerxen-dev/OnboardCI/issues/1)).
