package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type EnvStore struct {
	Paths Paths
}

func NewEnvStore(paths Paths) EnvStore {
	return EnvStore{Paths: paths}
}

func (s EnvStore) List() ([]string, error) {
	if err := s.Paths.Ensure(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.Paths.EnvsDir)
	if err != nil {
		return nil, fmt.Errorf("read envs: %w", err)
	}

	envs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".enc") {
			envs = append(envs, strings.TrimSuffix(name, ".enc"))
		}
	}
	sort.Strings(envs)
	return envs, nil
}

func (s EnvStore) EnvExists(env string) bool {
	info, err := os.Stat(s.Paths.EnvFile(env))
	return err == nil && !info.IsDir()
}

func (s EnvStore) KeyExists(env string) bool {
	info, err := os.Stat(s.Paths.KeyFile(env))
	return err == nil && !info.IsDir()
}

func (s EnvStore) RequireRunnableEnv(env string) error {
	if !s.EnvExists(env) {
		return fmt.Errorf("env %q does not exist", env)
	}
	if !s.KeyExists(env) {
		return fmt.Errorf("key for env %q does not exist", env)
	}
	return nil
}

func (s EnvStore) ReadCurrent() (string, error) {
	data, err := os.ReadFile(s.Paths.CurrentFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("current env is not set")
		}
		return "", fmt.Errorf("read current env: %w", err)
	}

	env := strings.TrimSpace(string(data))
	if env == "" {
		return "", fmt.Errorf("current env is not set")
	}
	if err := ValidateEnvName(env); err != nil {
		return "", fmt.Errorf("current env is invalid: %w", err)
	}
	return env, nil
}

func (s EnvStore) WriteCurrent(env string) error {
	if err := s.Paths.Ensure(); err != nil {
		return err
	}
	if err := os.WriteFile(s.Paths.CurrentFile, []byte(env+"\n"), 0600); err != nil {
		return fmt.Errorf("write current env: %w", err)
	}
	return ChmodPrivateFile(s.Paths.CurrentFile)
}

func (s EnvStore) RemoveTempFile(path string) {
	if path == "" {
		return
	}
	_ = os.Remove(path)
}

func SameBasePath(dir, env string) string {
	return filepath.Join(dir, env+".enc")
}
