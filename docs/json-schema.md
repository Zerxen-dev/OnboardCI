# OnboardCI JSON Output Schema

This document defines the stable JSON contract emitted by `onboardci` when using
the `--json` reporter. Tool integrations (CI pipelines, GitHub Action comment
bots, dashboard parsers) can rely on these fields.

## Top-level: `ExecutionReport`

| Field              | Type             | Required | Description |
|--------------------|------------------|----------|-------------|
| `success`          | boolean          | yes      | `true` if every step passed. |
| `total_commands`   | integer          | yes      | Number of steps extracted and attempted. |
| `passed_count`     | integer          | yes      | Number of steps that exited 0. |
| `failed_step`      | CommandResult    | no       | The first failing step (omitted when all pass). |
| `steps`            | CommandResult[]  | yes      | Per-step results, in execution order. |
| `executed_in_docker` | boolean        | yes      | Whether steps ran inside a Docker container. |

## Per-step: `CommandResult`

| Field             | Type    | Required | Description |
|-------------------|---------|----------|-------------|
| `source`          | string  | yes      | Source file the step came from (alias of `source_file`). |
| `source_file`     | string  | yes      | Source file the step came from. |
| `line`            | integer | yes      | Line number in the source file (alias of `line_number`). |
| `line_number`     | integer | yes      | Line number in the source file. |
| `command`         | string  | yes      | The command that was executed. |
| `step_type`       | string  | yes      | Step type (e.g. `test`, `setup`). |
| `status`          | string  | yes      | `passed` or `failed`. |
| `exit_code`       | integer | yes      | Process exit code. |
| `reason`          | string  | no       | Failure reason (alias of `failure_reason`). |
| `failure_reason`  | string  | no       | Failure reason (omitted on success). |
| `suggestion`      | string  | no       | Suggested fix (omitted on success). |
| `stdout`          | string  | no       | Captured stdout (omitted when empty). |
| `stderr`          | string  | no       | Captured stderr (omitted when empty). |
| `duration_ms`     | integer | yes      | Step duration in milliseconds. |

## Example payload

```json
{
  "success": false,
  "total_commands": 2,
  "passed_count": 1,
  "executed_in_docker": true,
  "failed_step": {
    "command": "npm test",
    "step_type": "test",
    "source_file": "README.md",
    "line_number": 42,
    "status": "failed",
    "exit_code": 1,
    "failure_reason": "missing env var DATABASE_URL",
    "suggestion": "document required env vars or provide .env.example"
  },
  "steps": [
    {
      "command": "npm ci",
      "step_type": "setup",
      "source_file": "README.md",
      "line_number": 10,
      "status": "passed",
      "exit_code": 0,
      "duration_ms": 1234
    },
    {
      "command": "npm test",
      "step_type": "test",
      "source_file": "README.md",
      "line_number": 42,
      "status": "failed",
      "exit_code": 1,
      "failure_reason": "missing env var DATABASE_URL",
      "suggestion": "document required env vars or provide .env.example",
      "duration_ms": 567
    }
  ]
}
```

## Backwards compatibility

The top-level summary keys `success`, `total_commands`, and `passed_count` are
guaranteed to remain. New fields may be added; existing fields will not be
renamed or removed without a major version bump.