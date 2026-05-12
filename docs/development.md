# dtx Developer Manual

This document summarizes the procedures for developers who work on and test dtx itself.

For installation and basic usage instructions for end users, see `README.md`.

## Prerequisites

### Required

* Go 1.21 or later
* dotenvx CLI

Check whether the dotenvx CLI is installed:

```bash
command -v dotenvx
```

If it is not installed:

```bash
npm install -g @dotenvx/dotenvx
```

## Basic Development Loop

After a normal change, verify it in the following order.

```bash
gofmt -w cmd internal
go test ./...
go build -o /private/tmp/dtx ./cmd/dtx
```

Notes:

* Do not generate the `dtx` binary in the repository root.
* Write build artifacts to `/private/tmp/dtx` or `/tmp/dtx`.
* Use `DTX_HOME` for manual checks so you do not pollute the real user's `~/.dtx`.

## Testing

Run all tests.

```bash
go test ./...
```

Run only a specific package.

```bash
go test ./internal/cli
go test ./internal/core
```

Show verbose logs.

```bash
go test -v ./...
```

## Build

Build a binary for development verification.

```bash
go build -o /private/tmp/dtx ./cmd/dtx
```

Quick version and help checks:

```bash
/private/tmp/dtx --help
/private/tmp/dtx --version
```

## Isolated Manual Verification

Point `DTX_HOME` at a temporary directory so you can verify behavior without modifying the real user's `~/.dtx`.

```bash
tmp_home=$(mktemp -d /private/tmp/dtx-home.XXXXXX)
DTX_HOME="$tmp_home" /private/tmp/dtx ls
```

## End-to-End Verification with Real dotenvx

Verify the flow from `edit` through `run` using the real dotenvx CLI.

```bash
go build -o /private/tmp/dtx ./cmd/dtx

tmp_home=$(mktemp -d /private/tmp/dtx-home.XXXXXX)
editor=$(mktemp /private/tmp/dtx-editor.XXXXXX)

cat > "$editor" <<'SH'
#!/bin/sh
printf 'HELLO=dev\n' > "$1"
SH
chmod +x "$editor"

DTX_HOME="$tmp_home" VISUAL="$editor" /private/tmp/dtx edit dev
DTX_HOME="$tmp_home" /private/tmp/dtx ls
DTX_HOME="$tmp_home" /private/tmp/dtx use dev
DTX_HOME="$tmp_home" /private/tmp/dtx current
DTX_HOME="$tmp_home" /private/tmp/dtx run dev -- sh -c 'printf "$HELLO\n"'
```

Expected output:

```text
Edited env: dev
dev
Using env: dev
dev
Using env: dev
dev
```

## Directory Layout

Primary implementation locations:

```text
cmd/dtx/main.go        CLI entry point
internal/cli/          Argument parsing and command dispatch
internal/command/      Implementation of each dtx command
internal/core/         Paths, env names, permissions, env store
internal/dotenvx/      dotenvx CLI adapter
internal/process/      Subprocess execution
internal/apperr/       Exit codes and error formatting
```

## dotenvx Integration Checkpoints

Keep dotenvx CLI calls encapsulated inside `internal/dotenvx`.

Things to verify:

* dtx itself must not interpret the dotenvx encryption format.
* Calls equivalent to `run` / `encrypt` / `decrypt` must go through the adapter.
* Use `--quiet` by default, and allow detailed output only with `--verbose`.
* Store env files in `envs/<env>.enc` and key files in `keys/<env>`.

## Permission Checks

After manual verification, check file permissions if needed.

```bash
find "$tmp_home" -maxdepth 2 -type d -exec ls -ld {} \;
find "$tmp_home" -maxdepth 2 -type f -exec ls -l {} \;
```

Expected values:

* dtx home, `envs/`, and `keys/` should be `700`.
* `current`, env files, and key files should be `600`.

## Frequently Used Verification Commands

Current diff:

```bash
git status --short
git diff
```

List Go files:

```bash
find cmd internal -type f | sort
```

Check the dotenvx CLI specification:

```bash
dotenvx run --help
dotenvx encrypt --help
dotenvx decrypt --help
```

## Release

Releases are automated through GitHub Actions + GoReleaser. Pushing a version tag is enough.

### Release Procedure

```bash
# 1. Confirm that main is up to date
git checkout main
git pull

# 2. Confirm that all tests pass
go test ./...

# 3. Tag the version and push it (this alone triggers the automated release)
git tag v0.1.0
git push origin v0.1.0
```

After the tag is pushed, `.github/workflows/release.yml` runs and automatically creates the following assets in the GitHub Release.

| Artifact | Contents |
|---|---|
| `dtx_linux_amd64.tar.gz` | Linux (x86_64) binary |
| `dtx_linux_arm64.tar.gz` | Linux (ARM64) binary |
| `dtx_darwin_amd64.tar.gz` | macOS (Intel) binary |
| `dtx_darwin_arm64.tar.gz` | macOS (Apple Silicon) binary |
| `dtx_windows_amd64.zip` | Windows (x86_64) binary |
| `checksums.txt` | SHA-256 checksums for each archive |

### Versioning Rules

Follow [Semantic Versioning](https://semver.org/).

* `v1.2.3` - official release published as a GitHub Release
* A suffix such as `v1.2.3-beta.1` -> automatically classified as a pre-release

### Checking the Version String

The release binary embeds the version using `ldflags`.

```bash
dtx --version  # -> dtx v0.1.0
```

A development build created with plain `go build` shows `dtx dev`.

### Verifying a Release Build Locally

If GoReleaser is installed, you can verify the artifacts locally without publishing to GitHub.

```bash
goreleaser release --snapshot --clean
# Binaries for each platform are generated in the dist/ directory
```

## Notes

* `dtx edit` uses `$VISUAL` first, then `$EDITOR`.
* If neither is set, `dtx edit` must return an error.
* `dtx run` only interprets dtx-side options before `--`.
* Everything after `--` is passed through to the target command as-is.
* `dtx doctor`, shell integration, and key rotation are out of scope for the MVP.
