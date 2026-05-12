# dtx

`dtx` is a small CLI for managing encrypted `.env`-style environments on your local machine.

It is built on top of the `dotenvx` CLI and gives you a narrow workflow:

* keep env files encrypted at rest
* select the env you want to use
* run commands through an explicit execution gate

The goal is not to replace your shell environment management. The goal is to make secret-backed command execution more deliberate and easier to reason about.

## Why dtx

With `dtx`, you can:

* keep multiple local environments such as `dev`, `staging`, and `prod`
* avoid leaving plaintext env files around by default
* avoid accidentally running commands with the wrong secrets loaded
* keep the "current env" as lightweight state without storing secrets in it

The intended flow looks like this:

```bash
dtx edit dev
dtx use dev
dtx run -- npm start
```

## Requirements

`dtx` depends on the `dotenvx` CLI for encryption, decryption, and runtime env injection.

Install `dotenvx` first:

```bash
npm install -g @dotenvx/dotenvx
```

Then install `dtx`:

```bash
go install github.com/yokawasa/dtx/cmd/dtx@latest
```

Requirements:

* Go 1.21 or later to install from source with `go install`
* `dotenvx` available on `PATH`

## Quick Start

Create or edit an env named `dev`:

```bash
dtx edit dev
```

`dtx edit` uses `$VISUAL` first, then `$EDITOR`. If neither is set, it returns an error.

Select the current env:

```bash
dtx use dev
```

Check the current env:

```bash
dtx current
```

Run a command with the selected env:

```bash
dtx run -- npm start
```

Run a command with an explicit env instead of `current`:

```bash
dtx run prod -- npm start
```

List available envs:

```bash
dtx ls
```

## Command Summary

```text
dtx edit <env>              Create or edit an encrypted env
dtx use <env>               Set the current env
dtx current                 Print the current env
dtx ls                      List available envs
dtx run [env] -- <command>  Run a command with an env
```

Notes:

* `dtx run` only interprets dtx options before `--`.
* Everything after `--` is passed to the target command unchanged.
* If no env is passed to `dtx run`, `dtx` uses the current env.

## Storage Layout

By default, `dtx` stores its state under `~/.dtx`.

```text
~/.dtx/
  envs/
    dev.enc
    prod.enc
  keys/
    dev
    prod
  current
```

`current` stores only the selected env name. It does not store secret values.

For isolated testing or automation, you can override the home directory with `DTX_HOME`:

```bash
DTX_HOME=/tmp/dtx-test dtx ls
```

## Security Model

`dtx` is designed to reduce mistakes and limit plaintext exposure, not to provide perfect isolation from the same local user.

In practice, that means:

* env files are stored encrypted
* plaintext is exposed only during editing or command execution
* `dtx run` is the explicit path for secret-backed execution
* the selected env state is stored separately from the secret data

## Design and Development Docs

If you want the implementation details or design rationale:

* [Design Document](docs/design.md)
* [Implementation Notes](docs/implementation.md)
* [Developer Manual](docs/development.md)
* [ADR 0001: Go and dotenvx CLI adapter](docs/adr/0001-go-with-dotenvx-cli-adapter.md)
