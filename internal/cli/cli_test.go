package cli

import (
	"bytes"
	"os"
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
	if err == nil || !strings.Contains(err.Error(), "command failed") {
		t.Fatalf("expected command failure, got %v", err)
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
	if _, err := os.Stat(filepath.Join(home, "envs", "empty.enc")); !os.IsNotExist(err) {
		t.Fatalf("env file exists after failed edit: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, "keys", "empty")); !os.IsNotExist(err) {
		t.Fatalf("key file exists after failed edit: %v", err)
	}
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
