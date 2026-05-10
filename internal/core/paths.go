package core

import (
	"fmt"
	"os"
	"path/filepath"
)

type Paths struct {
	Home        string
	EnvsDir     string
	KeysDir     string
	CurrentFile string
}

func NewPaths() (Paths, error) {
	home := os.Getenv("DTX_HOME")
	if home == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return Paths{}, fmt.Errorf("resolve user home: %w", err)
		}
		home = filepath.Join(userHome, ".dtx")
	}
	return Paths{
		Home:        home,
		EnvsDir:     filepath.Join(home, "envs"),
		KeysDir:     filepath.Join(home, "keys"),
		CurrentFile: filepath.Join(home, "current"),
	}, nil
}

func (p Paths) Ensure() error {
	for _, dir := range []string{p.Home, p.EnvsDir, p.KeysDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
		if err := os.Chmod(dir, 0700); err != nil {
			return fmt.Errorf("chmod %s: %w", dir, err)
		}
	}
	return nil
}

func (p Paths) EnvFile(env string) string {
	return filepath.Join(p.EnvsDir, env+".enc")
}

func (p Paths) KeyFile(env string) string {
	return filepath.Join(p.KeysDir, env)
}
