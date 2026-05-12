package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadCurrentReturnsErrorWhenEnvIsMissing(t *testing.T) {
	home := t.TempDir()
	paths := Paths{
		Home:        home,
		EnvsDir:     filepath.Join(home, "envs"),
		KeysDir:     filepath.Join(home, "keys"),
		CurrentFile: filepath.Join(home, "current"),
	}
	store := NewEnvStore(paths)

	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.CurrentFile, []byte("dev\n"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := store.ReadCurrent()
	if err == nil {
		t.Fatal("expected missing current env error")
	}
	if got, want := err.Error(), `current env "dev" does not exist`; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestReadCurrentReturnsEnvWhenEnvExists(t *testing.T) {
	home := t.TempDir()
	paths := Paths{
		Home:        home,
		EnvsDir:     filepath.Join(home, "envs"),
		KeysDir:     filepath.Join(home, "keys"),
		CurrentFile: filepath.Join(home, "current"),
	}
	store := NewEnvStore(paths)

	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.EnvFile("dev"), []byte("encrypted"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := store.WriteCurrent("dev"); err != nil {
		t.Fatal(err)
	}

	got, err := store.ReadCurrent()
	if err != nil {
		t.Fatalf("ReadCurrent failed: %v", err)
	}
	if got != "dev" {
		t.Fatalf("ReadCurrent = %q, want dev", got)
	}
}
