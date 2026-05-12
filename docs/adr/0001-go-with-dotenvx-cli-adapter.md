# ADR 0001: Implement with Go and a dotenvx CLI Adapter

## Status

Accepted

## Context

dtx is a CLI tool for safely managing environment variables in a local environment and for explicitly using the selected env through `dtx run`.

By design, the primary responsibilities of dtx are:

* Manage the selected env state.
* Manage encrypted env files and keys under `~/.dtx`.
* Restrict the execution path to `dtx run` to prevent running commands in the wrong environment.
* Delegate encryption and decryption behavior to dotenvx.

At first, a Node.js + TypeScript implementation using the dotenvx library API was considered. However, for a CLI tool, Go is a better fit for distribution, single-binary delivery, and straightforward handling of process execution and file permissions.

At the same time, dotenvx is a tool in the Node.js ecosystem, so calling its library API directly from Go is not a natural fit. Reimplementing dotenvx-compatible encryption and decryption in Go would also force dtx to track an encryption format that is outside its intended scope.

## Decision

Implement dtx itself in Go.

Do not reimplement encryption, decryption, or runtime injection in Go. Use the dotenvx CLI as a subprocess instead.

Do not scatter dotenvx CLI calls directly throughout dtx commands. Keep them inside an adapter layer.

Specifically, adopt the following:

* Implement dtx as a Go module.
* Use `cmd/dtx/main.go` as the CLI entry point.
* Treat the dotenvx CLI as a required external dependency.
* Place the adapter under `internal/dotenvx`.
* Let the adapter handle dotenvx CLI calls equivalent to `run` / `encrypt` / `decrypt`.
* Do not let dtx itself interpret the dotenvx encryption format.
* Keep dtx responsible for env selection, path resolution, key management, permission handling, output control, and error formatting.

## Consequences

### Positive

* dtx is easy to distribute as a single binary.
* Users do not need a Node.js runtime.
* Go's standard capabilities are well suited to file permissions, process execution, and exit-code propagation.
* dtx can stay focused on env management and the execution gate.
* dtx does not need to reimplement dotenvx encryption behavior.
* If the implementation is replaced later, the adapter layer provides a clean boundary.

### Negative

* The dotenvx CLI must be installed in the runtime environment.
* Changes to dotenvx CLI behavior or output can affect dtx.
* A subprocess-based integration gives less control than a library API.
* The design must carefully separate dotenvx-originated stdout from the target command's stdout.

### Neutral / Mitigation

* Check for the existence of the dotenvx CLI before executing commands that require it.
* Keep dotenvx CLI calls inside the adapter layer so the impact of specification changes stays localized.
* Show detailed dotenvx output only when `--verbose` is specified.
* Suppress output in normal operation using behavior equivalent to `--quiet`.
* Format dtx error messages in dtx terms instead of exposing raw dotenvx errors directly.

## Alternatives Considered

### Node.js + TypeScript + dotenvx library API

This is a natural integration with dotenvx, but it would require the dtx runtime to depend on Node.js.

It is not chosen because Go is a better fit for CLI distribution, single-binary delivery, and file-permission and process-control handling.

### Reimplement dotenvx-compatible encryption in Go

This would remove the external CLI dependency, but it would require following dotenvx encryption and key-management behavior.

That is not the core purpose of dtx, which is env management and an execution gate, so it is not chosen for the MVP.

### Call the dotenvx CLI Directly from Each Command in Go

This would shorten the implementation, but it would spread the dotenvx CLI dependency across the codebase.

That would make future replacement and testing harder, so the chosen approach keeps it inside an adapter layer.

## Follow-ups

* Finalize the concrete way to pass the private key at `~/.dtx/keys/<env>` to the dotenvx CLI during implementation.
* Measure the quiet/verbose behavior of `dotenvx run` and reflect it in the dtx output-control specification.
* Separate adapter tests that use the real dotenvx CLI from tests that use a fake adapter.
