package process

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunEditorExecsDirectlyWithPathAsSingleArg(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses POSIX shell scripts")
	}

	binDir := t.TempDir()
	editorPath := filepath.Join(binDir, "editor")
	argsFile := filepath.Join(t.TempDir(), "args")
	script := `#!/bin/sh
printf '%s\n' "$@" > "$DTX_EDITOR_ARGS_FILE"
`
	if err := os.WriteFile(editorPath, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DTX_EDITOR_ARGS_FILE", argsFile)

	envPath := filepath.Join(t.TempDir(), "env file with spaces.enc")
	err := Runner{}.RunEditor(editorPath+" --flag", envPath)
	if err != nil {
		t.Fatalf("RunEditor failed: %v", err)
	}

	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	want := []string{"--flag", envPath}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("editor args = %q, want %q", got, want)
	}
}

func TestRunEditorRejectsBlankEditor(t *testing.T) {
	err := Runner{}.RunEditor(" \t\n ", "env.enc")
	if err == nil || err.Error() != "editor is not set" {
		t.Fatalf("RunEditor error = %v, want editor is not set", err)
	}
}
