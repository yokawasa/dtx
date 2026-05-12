# dtx Implementation Notes

This document organizes the assumptions, MVP scope, and implementation policy used to turn `docs/design.md` into code.

## 1. Implementation Policy

### 1.1 Implementation Language

The MVP is implemented in Go.

Reasons:

* It is easy to distribute as a single-binary CLI tool.
* It handles file permissions, process execution, and exit-code handling well.
* It does not require a Node.js runtime in the user's environment.
* It fits local state management under `~/.dtx`.
* It is easy to test CLI behavior using temporary directories.

Notes:

* Do not reimplement dotenvx encryption and decryption behavior in Go.
* Treat dotenvx as a required external CLI dependency.
* Implement dtx itself in Go, and keep dotenvx CLI calls inside an adapter layer.

### 1.2 dotenvx Integration

In the Go implementation, use the dotenvx CLI as a subprocess rather than using a dotenvx library API.

Policy:

* The dotenvx CLI is a required dependency.
* Verify the presence of `dotenvx` at startup or before running commands that use it.
* Do not scatter direct dotenvx command calls throughout dtx. Keep them inside the `internal/dotenvx` adapter.
* Preserve a structure that can later be replaced with a Go-native implementation or another encryption backend.

### 1.3 Go Module

The MVP assumes a Go module.

Expected initialization:

```bash
go mod init github.com/yokawasa/dtx
```

The CLI entry point is `cmd/dtx/main.go`.

## 2. MVP Scope

Commands included in the MVP:

```bash
dtx use <env>
dtx current
dtx ls
dtx run [env] [--verbose] -- <command>
dtx edit <env>
```

Out of scope for the MVP:

* Key rotation
* Secret key injection via environment variables for CI
* OS keychain integration
* Passphrase-based mode
* Per-project `current`
* Automatic switching like `direnv`
* Built-in prompt rendering
* `dtx doctor`
* `dtx init shell`
* `dtx init completion`

## 3. Implementation Assumptions

### 3.1 `dtx edit <env>` Also Creates a New env

The design does not define a dedicated command for env creation.

For the MVP:

* `dtx edit <env>` creates a new env when the target env does not exist.
* When creating a new env, create an env-specific key at the same time.
* When the env already exists, reuse the existing key.

Reasons:

* The env creation flow can be completed without adding another command.
* This matches the intuition behind `edit`.

### 3.2 `dtx doctor` Is Out of Scope for the MVP

`docs/design.md` mentions `dtx doctor` in the context of checking key permissions, but it is not included in the CLI command list.

For the MVP:

* Do not implement `dtx doctor`.
* Perform only the minimum necessary permission checks when each command runs.
* Add `doctor` later as a separate diagnostic command.

### 3.3 Shell Integration Commands Are Out of Scope for the MVP

`dtx init shell` and `dtx init completion` are future additions.

For the MVP:

* Keep completion and prompt display as sample scripts only.
* Do not implement shell integration commands in dtx itself.

### 3.4 dotenvx CLI Adapter

The adapter owns the boundary between dtx and the dotenvx CLI.

Responsibilities:

* Check for the existence of the `dotenvx` command.
* Pass env file paths and keys.
* Handle calls equivalent to `run` / `encrypt` / `decrypt`.
* Control output equivalent to `--quiet` / `--verbose`.
* Convert dotenvx-originated errors into dtx-oriented errors.

Things to confirm during implementation:

* How to pass `~/.dtx/keys/<env>` to the dotenvx CLI
* Whether a temporary `.env.keys` file is required
* Whether `decrypt --stdout` / `encrypt --stdout` can be used
* Whether `dotenvx run` can suppress dotenvx logs while preserving the target command's stdout
* How to handle dotenvx CLI exit codes and error messages

Important constraints:

* dtx itself must not interpret the dotenvx encryption format.
* Details of dotenvx CLI invocation must not leak outside the adapter.

### 3.5 Position of `--verbose`

For the MVP, dtx options are interpreted only before `--`.

```bash
dtx run --verbose -- npm start
dtx run prod --verbose -- npm start
```

Everything after `--` is passed through to the target command as-is.

### 3.6 Confirmation for Dangerous envs

The design says confirmation should be possible for dangerous envs such as `prod`, but the concrete rules are undefined.

For the MVP:

* Do not implement an interactive confirmation prompt.
* Only print `Using env: prod`.
* Add future support for settings such as `confirm` or `protected env`.

### 3.7 Editor Selection

For the MVP:

* Prefer `$VISUAL`.
* Then use `$EDITOR`.
* Return an error if neither is set.

Reasons:

* Implicitly launching `vi` or something similar makes the experience vary by environment.
* This pushes users toward explicit editor configuration.

### 3.8 Switching Home for Tests

The MVP supports `DTX_HOME`.

* Use `~/.dtx` by default.
* If `DTX_HOME` is set, use that directory as the dtx home.

Example:

```bash
DTX_HOME=/tmp/dtx-test dtx ls
```

Reasons:

* Tests can run without polluting the real user's `~/.dtx`.
* It makes E2E tests easier to write.

## 4. Proposed File Layout

```text
go.mod
go.sum
cmd/
  dtx/
    main.go
internal/
  cli/
    cli.go
    parse.go
  command/
    use.go
    current.go
    ls.go
    run.go
    edit.go
  core/
    paths.go
    permissions.go
    env_name.go
    key_provider.go
    env_store.go
  dotenvx/
    adapter.go
    command.go
  process/
    runner.go
  output/
    output.go
  errors/
    errors.go
testdata/
```

## 5. Implementation Order

1. Create the Go module and CLI entry point.
2. Implement path management including `DTX_HOME` support.
3. Implement env-name validation and permission handling.
4. Implement `use` / `current` / `ls`.
5. Implement the dotenvx CLI adapter.
6. Implement `edit`.
7. Implement `run`.
8. Add error formatting and output control.
9. Add a minimal usage example to the README.
10. Add command-level tests.

## 6. Current Decisions

To keep implementation moving, the following are adopted as MVP assumptions.

* Implement dtx in Go.
* Use a Go module.
* Treat dotenvx as a required external CLI dependency.
* Keep dotenvx CLI calls inside an adapter layer.
* Let `edit` handle both editing an existing env and creating a new one.
* Do not implement `doctor` in the MVP.
* Do not implement `init shell` / `init completion` in the MVP.
* `run` launches the dotenvx CLI using Go process execution and propagates the exit code.
* Interpret `--verbose` only before `--`.
* Do not implement confirmation prompts for dangerous envs in the MVP.
* Use only `$VISUAL` / `$EDITOR` for editor selection, and return an error if neither is set.
* Support `DTX_HOME` for testing.
