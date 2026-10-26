# OnboardCI Non-Goals

To maintain a focused, reliable, and lightweight developer tool, OnboardCI explicitly defines the following non-goals.

---

## 1. Not a General CI Replacement
OnboardCI is **not** an alternative to GitHub Actions, GitLab CI, CircleCI, or Jenkins.
- It does not orchestrate deployment pipelines, handle complex matrix builds across 20 OS versions, or manage production artifacts.
- It is solely designed to answer one question: **"Can a new contributor clone this repository and follow the documented setup and test steps to get a working environment?"**

---

## 2. Not a Vulnerability or Security Scanner
OnboardCI does **not** perform static application security testing (SAST), software composition analysis (SCA), or CVE vulnerability scans.
- It does not scan dependencies for known CVEs (like Snyk, Dependabot, or Grype).
- It runs repository instructions inside an isolated container to ensure reproducibility and documentation accuracy, not to assess threat vectors.

---

## 3. No Telemetry or Remote Execution
OnboardCI has **zero telemetry**.
- No usage metrics, tracking pixels, crash analytics, or phone-home requests are sent.
- All command extraction and container executions happen locally or in your designated CI container runner.
- The only network interactions permitted are explicit Docker image pulls initiated by your local Docker daemon.

---

## 4. No Remote Cloud Execution or Hosted Accounts
OnboardCI requires **no user accounts, API keys, or SaaS subscriptions**.
- It is distributed as a single standalone Go binary.
- All verification steps execute within your local Docker daemon or ephemeral GitHub Actions runner.
