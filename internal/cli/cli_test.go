package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/yokawasa/dtx/internal/apperr"
)

func TestUseCurrentAndList(t *testing.T) {
	home := t.TempDir()
	t.Setenv("DTX_HOME", home)

	envsDir := filepath.Join(home, "envs")
	if err := os.MkdirAll(envsDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envsDir, "prod.enc"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envsDir, "dev.enc"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Run([]string{"ls"}, nil, &out, &bytes.Buffer{}); err != nil {
		t.Fatalf("ls failed: %v", err)
	}
	if got := out.String(); got != "dev\nprod\n" {
		t.Fatalf("ls output = %q", got)
	}

	out.Reset()
	if err := Run([]string{"use", "dev"}, nil, &out, &bytes.Buffer{}); err != nil {
		t.Fatalf("use failed: %v", err)
	}
	if got := out.String(); got != "Using env: dev\n" {
		t.Fatalf("use output = %q", got)
	}

	out.Reset()
	if err := Run([]string{"current"}, nil, &out, &bytes.Buffer{}); err != nil {
		t.Fatalf("current failed: %v", err)
	}
	if got := out.String(); got != "dev\n" {
		t.Fatalf("current output = %q", got)
	}
}

func TestRunUsesFakeDotenvxAndPropagatesExitCode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses POSIX shell scripts")
	}

	home := t.TempDir()
	t.Setenv("DTX_HOME", home)
	installFakeDotenvx(t)

	if err := os.MkdirAll(filepath.Join(home, "envs"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "keys"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "envs", "dev.enc"), []byte("HELLO=encrypted\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "keys", "dev"), []byte("KEY=x\n"), 0600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := Run([]string{"run", "dev", "--", "sh", "-c", "printf env=$DTX_FAKE_ENV"}, nil, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if got := out.String(); got != "Using env: dev\nenv=dev" {
		t.Fatalf("run output = %q", got)
	}

	err = Run([]string{"run", "dev", "--", "sh", "-c", "exit 7"}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected command failure, got nil")
	}
	if !apperr.IsSilent(err) {
		t.Fatalf("expected silent error, got %v", err)
	}
	if got := apperr.ExitCode(err); got != 7 {
		t.Fatalf("exit code = %d, want 7", got)
	}
}

func TestEditCreatesEnvWithFakeDotenvx(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses POSIX shell scripts")
	}

	home := t.TempDir()
	t.Setenv("DTX_HOME", home)
	installFakeDotenvx(t)
	editorPath := installEditor(t, "printf 'HELLO=world\\n' > \"$1\"\n")
	t.Setenv("VISUAL", editorPath)

	var out bytes.Buffer
	if err := Run([]string{"edit", "dev"}, nil, &out, &bytes.Buffer{}); err != nil {
		t.Fatalf("edit failed: %v", err)
	}
	if got := out.String(); got != "Edited env: dev\n" {
		t.Fatalf("edit output = %q", got)
	}

	envData, err := os.ReadFile(filepath.Join(home, "envs", "dev.enc"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(envData), "ENCRYPTED_BY_FAKE_DOTENVX=1") {
		t.Fatalf("env file was not encrypted by fake dotenvx: %q", string(envData))
	}
	if _, err := os.Stat(filepath.Join(home, "keys", "dev")); err != nil {
		t.Fatalf("key file was not created: %v", err)
	}
}

func TestEditDoesNotLeavePartialStateWhenNewEnvHasNoVariables(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses POSIX shell scripts")
	}

	home := t.TempDir()
	t.Setenv("DTX_HOME", home)
	installFakeDotenvx(t)
	editorPath := installEditor(t, "printf '# no variables\\n' > \"$1\"\n")
	t.Setenv("VISUAL", editorPath)

	err := Run([]string{"edit", "empty"}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected edit to fail")
	}
	if !strings.Contains(err.Error(), `env "empty" does not contain variables to encrypt`) {
		t.Fatalf("unexpected error: %v", err)
	}
	tmpDir := keptTempDir(t, err)
	if !strings.HasPrefix(tmpDir, home+string(os.PathSeparator)) {
		t.Fatalf("temporary edit directory = %q, want under %q", tmpDir, home)
	}
	data, readErr := os.ReadFile(filepath.Join(tmpDir, "empty.enc"))
	if readErr != nil {
		t.Fatalf("temporary edit file was not kept: %v", readErr)
	}
	if got := string(data); got != "# no variables\n" {
		t.Fatalf("temporary edit file = %q", got)
	}
	if _, err := os.Stat(filepath.Join(home, "envs", "empty.enc")); !os.IsNotExist(err) {
		t.Fatalf("env file exists after failed edit: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, "keys", "empty")); !os.IsNotExist(err) {
		t.Fatalf("key file exists after failed edit: %v", err)
	}
}

func keptTempDir(t *testing.T, err error) string {
	t.Helper()

	const marker = "temporary edit directory kept at "
	index := strings.LastIndex(err.Error(), marker)
	if index == -1 {
		t.Fatalf("error does not include kept temp dir: %v", err)
	}
	return err.Error()[index+len(marker):]
}

func TestEditUsesTemporaryKeyUnderDtxHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses POSIX shell scripts")
	}

	home := t.TempDir()
	t.Setenv("DTX_HOME", home)
	installFakeDotenvx(t)
	keyLog := filepath.Join(t.TempDir(), "key-path")
	t.Setenv("DTX_FAKE_KEY_LOG", keyLog)
	editorPath := installEditor(t, "printf 'HELLO=world\\n' > \"$1\"\n")
	t.Setenv("VISUAL", editorPath)

	if err := Run([]string{"edit", "dev"}, nil, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatalf("edit failed: %v", err)
	}

	data, err := os.ReadFile(keyLog)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(data))
	if !strings.HasPrefix(got, home+string(os.PathSeparator)) {
		t.Fatalf("temporary key path = %q, want under %q", got, home)
	}
	if got == filepath.Join(home, "keys", "dev") {
		t.Fatalf("encrypt used target key path directly: %q", got)
	}
}

func TestUseMissingEnvReturnsError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("DTX_HOME", home)
	if err := os.MkdirAll(filepath.Join(home, "envs"), 0700); err != nil {
		t.Fatal(err)
	}

	err := Run([]string{"use", "missing"}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), `"missing" does not exist`) {
		t.Fatalf("expected 'does not exist' error, got %v", err)
	}
}

func TestCurrentWithNoCurrentFileReturnsNotSet(t *testing.T) {
	home := t.TempDir()
	t.Setenv("DTX_HOME", home)
	if err := os.MkdirAll(filepath.Join(home, "envs"), 0700); err != nil {
		t.Fatal(err)
	}

	err := Run([]string{"current"}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "current env is not set") {
		t.Fatalf("expected 'current env is not set' error, got %v", err)
	}
}

func TestInvalidEnvNamesAreRejected(t *testing.T) {
	home := t.TempDir()
	t.Setenv("DTX_HOME", home)

	cases := []struct {
		name string
		args []string
	}{
		{"use ../etc", []string{"use", "../etc"}},
		{"use prod/key", []string{"use", "prod/key"}},
		{"run ../etc", []string{"run", "../etc", "--", "sh", "-c", "true"}},
		{"run prod/key", []string{"run", "prod/key", "--", "sh", "-c", "true"}},
		{"edit ../etc", []string{"edit", "../etc"}},
		{"edit prod/key", []string{"edit", "prod/key"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Run(tc.args, nil, &bytes.Buffer{}, &bytes.Buffer{})
			if err == nil || !strings.Contains(err.Error(), "invalid env name") {
				t.Fatalf("expected 'invalid env name' error, got %v", err)
			}
		})
	}
}

func TestUsageErrorsPrintPlainUsageAndExitSilently(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStderr string
	}{
		{
			name:       "no args",
			args:       nil,
			wantStderr: usage,
		},
		{
			name:       "use args",
			args:       []string{"use"},
			wantStderr: "usage: dtx use <env>\n",
		},
		{
			name:       "current args",
			args:       []string{"current", "extra"},
			wantStderr: "usage: dtx current\n",
		},
		{
			name:       "ls args",
			args:       []string{"ls", "extra"},
			wantStderr: "usage: dtx ls\n",
		},
		{
			name:       "run args",
			args:       []string{"run", "dev"},
			wantStderr: "usage: dtx run [env] [--verbose] -- <command>\n",
		},
		{
			name:       "edit args",
			args:       []string{"edit", "dev", "prod"},
			wantStderr: "usage: dtx edit <env> [--verbose]\n",
		},
		{
			name:       "completion args",
			args:       []string{"completion"},
			wantStderr: "usage: dtx completion <bash|zsh|fish>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			err := Run(tt.args, nil, &stdout, &stderr)
			if err == nil {
				t.Fatal("expected usage error")
			}
			if !apperr.IsSilent(err) {
				t.Fatalf("expected silent error, got %v", err)
			}
			if got := apperr.ExitCode(err); got != 2 {
				t.Fatalf("exit code = %d, want 2", got)
			}
			if got := stdout.String(); got != "" {
				t.Fatalf("stdout = %q", got)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", got, tt.wantStderr)
			}
		})
	}
}

func TestHelpPrintsFullUsageToStdout(t *testing.T) {
	wantContains := []string{
		"Usage:\n",
		"Commands:\n",
		"dtx run [env] [--verbose] -- <command>",
		"dtx completion <bash|zsh|fish>",
	}

	for _, flag := range []string{"-h", "--help"} {
		t.Run(flag, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			err := Run([]string{flag}, nil, &stdout, &stderr)
			if err != nil {
				t.Fatalf("help failed: %v", err)
			}
			got := stdout.String()
			if got == "" {
				t.Fatal("stdout is empty")
			}
			for _, want := range wantContains {
				if !strings.Contains(got, want) {
					t.Fatalf("stdout = %q, want substring %q", got, want)
				}
			}
			if got := stderr.String(); got != "" {
				t.Fatalf("stderr = %q", got)
			}
		})
	}
}

func TestCompletionOutputsShellScript(t *testing.T) {
	tests := []struct {
		shell        string
		wantContains []string
	}{
		{
			shell: "bash",
			wantContains: []string{
				"_dtx_envs()",
				"complete -o bashdefault -o default -F _dtx dtx",
				"use current ls run edit completion",
			},
		},
		{
			shell: "zsh",
			wantContains: []string{
				"#compdef dtx",
				"autoload -Uz compinit",
				"_dtx_env_names()",
				"compdef _dtx dtx",
				"completion:generate shell completion",
			},
		},
		{
			shell: "fish",
			wantContains: []string{
				"function __dtx_envs",
				"complete -c dtx -n '__fish_use_subcommand' -a 'use current ls run edit completion'",
				"complete -c dtx -n '__fish_seen_subcommand_from run; and __dtx_run_needs_separator' -a --",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			err := Run([]string{"completion", tt.shell}, nil, &stdout, &stderr)
			if err != nil {
				t.Fatalf("completion failed: %v", err)
			}
			if got := stderr.String(); got != "" {
				t.Fatalf("stderr = %q", got)
			}
			output := stdout.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Fatalf("completion output missing %q\noutput:\n%s", want, output)
				}
			}
		})
	}
}

func TestCompletionScriptsParseInTargetShells(t *testing.T) {
	tests := []struct {
		shell string
		args  []string
	}{
		{shell: "bash", args: []string{"-n"}},
		{shell: "zsh", args: []string{"-n"}},
		{shell: "fish", args: []string{"-n"}},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			shellPath, err := exec.LookPath(tt.shell)
			if err != nil {
				t.Skipf("%s not installed", tt.shell)
			}

			scriptPath := writeCompletionScript(t, tt.shell)
			cmdArgs := append(append([]string{}, tt.args...), scriptPath)
			cmd := exec.Command(shellPath, cmdArgs...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s parse failed: %v\n%s", tt.shell, err, output)
			}
		})
	}
}

func TestCompletionScriptsListEnvNames(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell completion integration tests require POSIX shells")
	}

	tests := []struct {
		shell   string
		command string
	}{
		{shell: "bash", command: `. "$1"; _dtx_envs`},
		{shell: "zsh", command: `source "$1"; _dtx_env_names`},
		{shell: "fish", command: `source $argv[1]; __dtx_envs`},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			shellPath, err := exec.LookPath(tt.shell)
			if err != nil {
				t.Skipf("%s not installed", tt.shell)
			}

			home := t.TempDir()
			envsDir := filepath.Join(home, "envs")
			if err := os.MkdirAll(envsDir, 0700); err != nil {
				t.Fatal(err)
			}
			for _, env := range []string{"prod", "dev"} {
				if err := os.WriteFile(filepath.Join(envsDir, env+".enc"), []byte("x"), 0600); err != nil {
					t.Fatal(err)
				}
			}

			scriptPath := writeCompletionScript(t, tt.shell)
			var cmd *exec.Cmd
			switch tt.shell {
			case "fish":
				cmd = exec.Command(shellPath, "-c", tt.command, scriptPath)
			default:
				cmd = exec.Command(shellPath, "-c", tt.command, "dtx-completion-test", scriptPath)
			}
			cmd.Env = append(os.Environ(), "DTX_HOME="+home, "HOME="+t.TempDir())
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s env listing failed: %v\n%s", tt.shell, err, output)
			}
			if got := strings.TrimSpace(string(output)); got != "dev\nprod" {
				t.Fatalf("%s env listing = %q, want %q", tt.shell, got, "dev\nprod")
			}
		})
	}
}

func writeCompletionScript(t *testing.T, shell string) string {
	t.Helper()

	var stdout bytes.Buffer
	if err := Run([]string{"completion", shell}, nil, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("completion failed: %v", err)
	}

	path := filepath.Join(t.TempDir(), "completion."+shell)
	if err := os.WriteFile(path, stdout.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func installFakeDotenvx(t *testing.T) {
	t.Helper()
	binDir := t.TempDir()
	path := filepath.Join(binDir, "dotenvx")
	script := `#!/bin/sh
if [ "$1" = "--quiet" ] || [ "$1" = "--verbose" ]; then
  shift
fi
cmd="$1"
shift
case "$cmd" in
  run)
    while [ "$1" != "--" ]; do
      shift
    done
    shift
    DTX_FAKE_ENV=dev exec "$@"
    ;;
  decrypt)
    printf 'HELLO=world\n'
    ;;
  encrypt)
    file=""
    key=""
    while [ "$#" -gt 0 ]; do
      case "$1" in
        -f) shift; file="$1" ;;
        -fk) shift; key="$1" ;;
      esac
      shift
    done
    if [ -n "$DTX_FAKE_KEY_LOG" ]; then
      printf '%s\n' "$key" > "$DTX_FAKE_KEY_LOG"
    fi
    if grep -Eq '^[A-Za-z_][A-Za-z0-9_]*=' "$file"; then
      printf '\nENCRYPTED_BY_FAKE_DOTENVX=1\n' >> "$file"
      mkdir -p "$(dirname "$key")"
      printf 'KEY=1\n' > "$key"
    fi
    ;;
esac
`
	if err := os.WriteFile(path, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func installEditor(t *testing.T, body string) string {
	t.Helper()
	binDir := t.TempDir()
	path := filepath.Join(binDir, "editor")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0700); err != nil {
		t.Fatal(err)
	}
	return path
}
