# dtx Design Document

## 1. Purpose

dtx is a CLI tool for safely and efficiently managing and switching environment variables in a local environment, including sensitive values such as API keys.

Its primary goals are:

* Store sensitive information securely.
* Switch between environments quickly.
* Prevent commands from running in the wrong environment.

---

## 2. Design Principles

### 2.1 Treat Each env as a Complete Unit

* Do not compose envs at runtime.
* Use only one env per execution.
* Each env must be self-contained.

---

### 2.2 Enforce a Single Execution Path

* Allow environment-variable-backed execution only through `dtx run`.
* Keep it fully separate from ordinary commands.

```bash
npm start       # without secrets
dtx run ...     # with secrets
```

---

### 2.3 Separate State from Execution

* `use` only selects state.
* `run` is the only execution gate.

---

### 2.4 Minimize Plaintext Exposure

* Keep envs encrypted at rest by default.
* Decrypt only temporarily during execution or editing.

---

## 3. Directory Structure

```text
~/.dtx/
  envs/
    dev.enc
    prod.enc
  current
```

### Role of Each Element

* `envs/`
  Stores encrypted env files.

* `current`
  Stores the currently selected env name only. It must not contain plaintext secrets.

---

### Permission Settings

* `~/.dtx`: `700`
* Each file: `600`

Only the owning user can access them.

---

## 4. CLI Specification

### 4.1 Command List

```bash
dtx use <env>
dtx current
dtx ls
dtx run [env] -- <command>
dtx edit <env>
```

---

### 4.2 Behavior of Each Command

#### `dtx use`

```bash
dtx use dev
```

* Only rewrites `current`.
* Does not inject environment variables.

---

#### `dtx run`

```bash
dtx run -- npm start
dtx run prod -- npm start
```

* Decrypts the env and executes the command.
* Uses `current` when no env is specified.
* Returns an error when `current` is not set.

Example output:

```text
Using env: dev
```

---

#### `dtx current`

```bash
dtx current
```

* Prints the current env name.

---

#### `dtx ls`

```bash
dtx ls
```

* Prints the list of available envs.

---

#### `dtx edit`

```bash
dtx edit dev
```

* Safely edits an env.

---

## 5. Execution Model

### 5.1 Execution Flow

When `dtx run` executes:

1. Determine the target env.
2. Read the decryption key for that env.
3. Pass the env file and decryption key to the dotenvx CLI adapter.
4. Execute the command through the dotenvx CLI.
5. Discard decrypted data.

---

### 5.2 Why This Execution Style Was Chosen

The implementation launches the dotenvx CLI as a subprocess from Go and executes the target command through that CLI.

* dtx does not keep decrypted env content in memory longer than necessary.
* dtx does not reimplement dotenvx encryption and decryption behavior.
* dtx can keep its responsibility limited to env selection and the execution gate.
* Exit codes and signals follow the result of the executed command.

---

## 6. Security Design

### 6.1 Separation

* Keep normal commands separate from env-backed execution.
* Make secret-backed execution unavailable unless the user explicitly goes through dtx.

---

### 6.2 Safety of State Management

* `current` stores only the env name.
* It never contains secret data.

---

### 6.3 Plaintext Handling

* Keep data encrypted by default.
* Decrypt only when needed.

---

### 6.4 Safety During Editing

Use a temporary-file workflow.

#### Flow

1. Create a temporary file with permission `600`.
2. Decrypt and write its contents.
3. Edit it in the editor.
4. Re-encrypt after saving.
5. Delete the temporary file.

#### Characteristics

* Plaintext is exposed only for a short period.
* Exposure stays within a manageable boundary.

---

## 7. Risks and Mitigations

### 7.1 Access by the Same User

* Full protection is not possible.
* The tool is primarily intended to prevent operational mistakes.

---

### 7.2 Running in the Wrong Environment

Mitigations:

* Show the env name during `dtx run`.
* Allow confirmation for dangerous envs such as `prod`.

---

### 7.3 Plaintext Leakage

Mitigations:

* Encrypted storage
* Temporary-file workflow
* Immediate deletion

---

## 8. Design Characteristics

* Predictable behavior
* Clear separation between state and execution responsibilities
* Intuitive as a CLI
* Minimal command set

## 9. Next Actions

The following three areas need more detailed design work.

### 1. Concrete Integration with dotenvx

dtx will not reimplement encryption itself. Instead, it will use dotenvx as the backend for encryption, decryption, and runtime injection.

#### Basic Policy

* dtx manages which env to use.
* dotenvx is responsible for how envs are encrypted and decrypted.
* `dtx run` determines the target env and then executes the command through dotenvx.
* dotenvx is a required dependency.
* Use the dotenvx CLI as a subprocess.
* Keep dotenvx CLI calls inside an adapter layer.

#### Handling env Files

The internal storage format is a dotenvx-compatible env file.

```text
~/.dtx/
  envs/
    dev.enc
    prod.enc
  current
  keys/
    dev
    prod
```

* `envs/<env>.enc` is an env file encrypted by dotenvx.
* `keys/<env>` stores the decryption key for that env.
* Each env name maps one-to-one to a file name.

The file extension is fixed as `.enc`.

#### `run` Integration

```bash
dtx run -- npm start
dtx run prod -- npm start
```

Internally, this is translated into the following flow:

1. Determine the target env.
2. Read the decryption key for that env.
3. Pass the env file and decryption key to the dotenvx CLI adapter.
4. Launch the dotenvx CLI as a subprocess.
5. Execute the command through the dotenvx CLI.

Conceptually:

```text
dtx
  -> resolve ~/.dtx/envs/dev.enc
  -> load ~/.dtx/keys/dev
  -> invoke dotenvx CLI through adapter
  -> dotenvx run -f ~/.dtx/envs/dev.enc -- npm start
```

From the dtx side, the goal is that users do not need to think about `DOTENV_PRIVATE_KEY` or `dotenvx run` directly. Details of the dotenvx CLI invocation stay inside the adapter layer.

#### Output Control

By default, stdout originating from dotenvx is not shown.

* Suppress it in normal operation using behavior equivalent to `--quiet`.
* Show dotenvx-originated stdout only for `dtx run --verbose ...`.
* Always show the target command's stdout and stderr normally.

This keeps normal dtx output minimal while still allowing dotenvx details to be inspected during troubleshooting.

#### `edit` Integration

```bash
dtx edit dev
```

Use a temporary file during editing.

1. Decrypt `envs/dev.enc` into a temporary file.
2. Edit it in the editor.
3. Re-encrypt it with dotenvx after saving.
4. Write the encrypted result back to `envs/dev.enc`.
5. Delete the temporary file.

At that point, the rule for whether to reuse an existing public/private key pair or issue a new key pair must be defined explicitly.

#### Error Design

Clearly distinguish the following cases:

* dotenvx is not found
* the env file does not exist
* the decryption key does not exist
* decryption failed
* the executed command failed

Because dotenvx is a required dependency, verify its availability at startup or initialization time.

dtx should not expose dotenvx errors verbatim. It should convert them into messages that make sense in dtx terms. When `--verbose` is specified, it may also show dotenvx-originated details.

```text
dtx: failed to decrypt env "prod"
dtx: dotenvx dependency is not available
dtx: current env is not set
```

#### Decisions

* dotenvx is a required dependency.
* Use the dotenvx CLI as a subprocess.
* Keep dotenvx CLI calls inside an adapter layer.
* Use `envs/<env>.enc` as the env file path.
* Do not show dotenvx-originated stdout by default.
* Show dotenvx-originated stdout only when `--verbose` is specified.

---

### 2. Encryption Key Management

Encryption keys must be protected more carefully than the env files themselves. In dtx, key storage is abstracted so the implementation can start with the simplest workable option.

#### Managed Objects

To match dotenvx's encryption model, handle at least the following:

* Public key
  A key used for encryption that may be included in the env file.

* Private key
  A key required for decryption.

The private key is the asset dtx must protect most carefully.

#### Candidate Storage Options

##### Local File Storage

```text
~/.dtx/
  keys/
    dev
    prod
```

* Simple to implement
* Easy to handle across platforms
* File permissions fixed to `600`
* Does not provide complete protection from the same user

For the MVP, this is the primary candidate.

##### Passphrase-Based Storage

Store the private key encrypted with an additional passphrase.

* Stronger protection if the private key file itself leaks
* Requires passphrase entry at runtime
* Poor fit for automation and scripting

Because it reduces CLI usability, it is not required in the initial implementation.

##### OS Keychain Storage

Use macOS Keychain, Windows Credential Manager, Linux Secret Service, and similar systems.

* Can use OS-level protection mechanisms
* Good user experience
* Large differences across operating systems
* Harder to handle in CI and headless environments

Consider this as a future option.

#### Recommended Policy

For the initial implementation, use the following policy:

1. Store private keys in `~/.dtx/keys/<env>`.
2. Enforce permission `600`.
3. Detect permission problems with `dtx doctor`.
4. Make it possible to swap the key provider in the future.

```text
KeyProvider
  file
  passphrase
  os-keychain
```

dtx itself should not depend on the storage mechanism. It should obtain private keys through `KeyProvider`.

#### Decisions

* Do not share the same key across multiple envs.
* Give each env its own key.
* Do not implement a key rotation command in the MVP.
* Do not implement an MVP mode that passes private keys through environment variables for CI.
* Treat key backup as outside the responsibility of dtx.

#### Future Work

* Add a command for key rotation.
* Add official support for passing private keys through environment variables for CI.

#### Handling Key Backups

dtx does not provide a key backup feature. If a private key is lost, the corresponding env file can no longer be decrypted.

Users should back up keys under `~/.dtx/keys/` on their own responsibility if needed.

---

### 3. Shell Integration

Shell integration is an auxiliary feature to reduce mistakes during normal use. Environment-variable injection remains limited to `dtx run`, and shell integration must not keep secrets resident in the current shell.

#### Completion

Target shells:

* zsh
* bash
* fish

Completion targets:

* subcommands
* env names
* `run` options

Examples:

```bash
dtx use <TAB>
dtx run <TAB>
```

Generate env-name completion from the contents of `~/.dtx/envs/`.

#### Prompt Display

Allow the currently selected env to be shown in the prompt.

Example:

```text
[dtx:dev] ~/app %
```

Only the env name stored in `current` is shown. Secret data and env contents are never read.

#### Shell Hooks

Limit shell hooks to the following purposes:

* Read the current env name when the prompt is rendered.
* Load completion functions.

Prohibited behavior:

* Do not decrypt env files when the shell starts.
* Do not keep secrets resident in the shell through `export`.
* Do not inject envs automatically just because the current directory changed.

#### Setup Commands

The following commands may be added in the future:

```bash
dtx init shell
dtx init completion
```

Example output:

```bash
eval "$(dtx init shell)"
```

Do not automatically append anything to shell configuration files. The default model is that users add it explicitly.

#### Decisions

* For prompt display, read `~/.dtx/current` directly on each render.
* Do not implement caching in the MVP.
* If `current` does not exist, show nothing.
* If reading `current` fails, do not fail prompt rendering; just hide the display.
* Do not provide prompt display from dtx itself.
* Keep prompt display as a sample script only.
* Do not support per-project `current`.
* Automatic switching like `direnv` is out of scope.
