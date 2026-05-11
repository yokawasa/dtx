package dotenvx

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestAdapterDoesNotPassNoOpsFlag(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses POSIX shell scripts")
	}

	argsFile := installFakeDotenvx(t)
	adapter := NewAdapter(nil, &bytes.Buffer{}, &bytes.Buffer{})

	tests := []struct {
		name string
		run  func() error
		want []string
	}{
		{
			name: "run",
			run: func() error {
				_, err := adapter.Run("env.enc", "key", []string{"printenv", "HELLO"}, false)
				return err
			},
			want: []string{"--quiet", "run", "--strict", "-f", "env.enc", "-fk", "key", "--", "printenv", "HELLO"},
		},
		{
			name: "decrypt",
			run: func() error {
				_, err := adapter.Decrypt("env.enc", "key", false)
				return err
			},
			want: []string{"--quiet", "decrypt", "--stdout", "-f", "env.enc", "-fk", "key"},
		},
		{
			name: "encrypt",
			run: func() error {
				return adapter.Encrypt("env.enc", "key", false)
			},
			want: []string{"--quiet", "encrypt", "-f", "env.enc", "-fk", "key"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); err != nil {
				t.Fatalf("adapter call failed: %v", err)
			}

			got := readArgs(t, argsFile)
			if contains(got, "--no-ops") {
				t.Fatalf("args include --no-ops: %v", got)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("args = %v, want %v", got, tt.want)
			}
		})
	}
}

func installFakeDotenvx(t *testing.T) string {
	t.Helper()

	binDir := t.TempDir()
	argsFile := filepath.Join(t.TempDir(), "args")
	path := filepath.Join(binDir, "dotenvx")
	script := `#!/bin/sh
: > "$DTX_DOTENVX_ARGS_FILE"
for arg do
  printf '%s\n' "$arg" >> "$DTX_DOTENVX_ARGS_FILE"
done

case "$2" in
  decrypt)
    printf 'HELLO=world\n'
    ;;
esac
`
	if err := os.WriteFile(path, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("DTX_DOTENVX_ARGS_FILE", argsFile)
	return argsFile
}

func readArgs(t *testing.T, path string) []string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Fields(string(data))
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
