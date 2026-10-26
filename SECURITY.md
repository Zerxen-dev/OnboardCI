# Security Policy

## 1. Supported Versions

Security updates are actively maintained for the latest release branch of OnboardCI.

| Version | Supported |
|---|---|
| `v0.1.x` / `main` | :white_check_mark: Yes |
| `< v0.1.0` | :x: No |

---

## 2. Execution Model & Sandbox Safety Notice

OnboardCI operates by extracting commands from documentation (`README.md`, `CONTRIBUTING.md`) or configuration (`.onboardci.yml`) and executing them.

**Please review the following security principles:**
- **Opt-in Execution**: Default invocation (`onboardci check`) is strictly a **dry-run** that only lists extracted commands. Command execution (`--execute`) is strictly **opt-in**.
- **Container Isolation**: Execution runs in a sandboxed Docker container without host bind mounts. The repository contents are copied into the container.
- **Resource Constraints**: By default, containers are bounded by CPU limits, memory limits, and non-root execution (`--user 1000:1000`).
- **Context Boundaries**: Running commands from untrusted pull requests should always be treated with caution. OnboardCI is not an arbitrary code sandbox for malicious payloads; it is an onboarding documentation verification tool.

---

## 3. Reporting a Vulnerability

If you identify a security vulnerability in OnboardCI (such as a container escape, command injection in flag handling, or unexpected privilege escalation):

1. **Do not create a public GitHub issue.**
2. Send an email to `security@zerxen.dev` with the subject `[SECURITY-ONBOARDCI] <Brief Description>`.
3. Include:
   - Reproduction steps or proof of concept
   - Operating system and Docker version
   - Potential impact
4. We acknowledge receipt within 48 hours and work with you to release a patch and advisory.
