# dtx

A minimal CLI tool for securely managing and switching encrypted environment variables, built on top of dotenvx.

## Requirements

`dtx` uses the `dotenvx` CLI for encryption, decryption, and runtime env injection.

```bash
npm install -g @dotenvx/dotenvx
```

## Usage

```bash
go install github.com/yokawasa/dtx/cmd/dtx@latest
```

Create or edit an env:

```bash
VISUAL="$EDITOR" dtx edit dev
```

Select the current env:

```bash
dtx use dev
dtx current
```

Run a command with the selected env:

```bash
dtx run -- npm start
```

Run with an explicit env:

```bash
dtx run prod -- npm start
```

List available envs:

```bash
dtx ls
```

For tests or isolated usage, set `DTX_HOME`:

```bash
DTX_HOME=/tmp/dtx-test dtx ls
```
