package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yokawasa/dtx/internal/core"
)

func Edit(ctx Context, env string, verbose bool) error {
	if err := validateEnvArg(env); err != nil {
		return err
	}
	if err := ctx.Paths.Ensure(); err != nil {
		return err
	}

	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		return fmt.Errorf("VISUAL or EDITOR is required for dtx edit")
	}

	tmpDir, err := os.MkdirTemp(ctx.Paths.Home, "edit-*")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	cleanupTmpDir := true
	defer func() {
		if cleanupTmpDir {
			_ = os.RemoveAll(tmpDir)
		}
	}()

	tmpEnvFile := core.SameBasePath(tmpDir, env)
	tmpKeyFile := filepath.Join(tmpDir, env+".key")
	targetEnvFile := ctx.Paths.EnvFile(env)
	targetKeyFile := ctx.Paths.KeyFile(env)
	isNewEnv := !ctx.Store.EnvExists(env)
	hadKey := ctx.Store.KeyExists(env)

	if !isNewEnv {
		if !hadKey {
			return fmt.Errorf("key for env %q does not exist", env)
		}
		if err := copyPrivateFile(targetKeyFile, tmpKeyFile); err != nil {
			return fmt.Errorf("copy temporary key file: %w", err)
		}
		plain, err := ctx.Dotenvx.Decrypt(targetEnvFile, tmpKeyFile, verbose)
		if err != nil {
			return fmt.Errorf("failed to decrypt env %q: %w", env, err)
		}
		if err := os.WriteFile(tmpEnvFile, plain, 0600); err != nil {
			return fmt.Errorf("write temporary env file: %w", err)
		}
	} else {
		if err := os.WriteFile(tmpEnvFile, []byte("# Add variables as KEY=value\n"), 0600); err != nil {
			return fmt.Errorf("write temporary env file: %w", err)
		}
	}
	if err := core.ChmodPrivateFile(tmpEnvFile); err != nil {
		return err
	}

	if err := ctx.Runner.RunEditor(editor, tmpEnvFile); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}
	cleanupTmpDir = false

	if err := ctx.Dotenvx.Encrypt(tmpEnvFile, tmpKeyFile, verbose); err != nil {
		return keepEditedTempDir(tmpDir, fmt.Errorf("failed to encrypt env %q: %w", env, err))
	}
	if !fileExists(tmpKeyFile) {
		return keepEditedTempDir(tmpDir, fmt.Errorf("env %q does not contain variables to encrypt", env))
	}

	if err := core.ChmodPrivateFile(tmpEnvFile); err != nil {
		return keepEditedTempDir(tmpDir, err)
	}
	if err := core.ChmodPrivateFile(tmpKeyFile); err != nil {
		return keepEditedTempDir(tmpDir, err)
	}

	if err := saveEditedEnv(tmpEnvFile, tmpKeyFile, targetEnvFile, targetKeyFile, !isNewEnv, hadKey); err != nil {
		return keepEditedTempDir(tmpDir, err)
	}
	cleanupTmpDir = true

	fmt.Fprintf(ctx.Stdout, "Edited env: %s\n", env)
	return nil
}

func keepEditedTempDir(tmpDir string, err error) error {
	return fmt.Errorf("%w; temporary edit directory kept at %s", err, tmpDir)
}

func saveEditedEnv(tmpEnvFile, tmpKeyFile, targetEnvFile, targetKeyFile string, hadEnv, hadKey bool) error {
	oldEnv, oldKey, err := readExistingTargets(targetEnvFile, targetKeyFile, hadEnv, hadKey)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(targetEnvFile), 0700); err != nil {
		return fmt.Errorf("create env directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(targetKeyFile), 0700); err != nil {
		return fmt.Errorf("create key directory: %w", err)
	}

	if err := os.Rename(tmpEnvFile, targetEnvFile); err != nil {
		return fmt.Errorf("save env file: %w", err)
	}
	if err := os.Rename(tmpKeyFile, targetKeyFile); err != nil {
		restoreEditedEnvTargets(targetEnvFile, targetKeyFile, oldEnv, oldKey, hadEnv, hadKey)
		return fmt.Errorf("save key file: %w", err)
	}
	return nil
}

func readExistingTargets(envFile, keyFile string, hadEnv, hadKey bool) ([]byte, []byte, error) {
	var oldEnv []byte
	var oldKey []byte
	var err error
	if hadEnv {
		oldEnv, err = os.ReadFile(envFile)
		if err != nil {
			return nil, nil, fmt.Errorf("read existing env file: %w", err)
		}
	}
	if hadKey {
		oldKey, err = os.ReadFile(keyFile)
		if err != nil {
			return nil, nil, fmt.Errorf("read existing key file: %w", err)
		}
	}
	return oldEnv, oldKey, nil
}

func restoreEditedEnvTargets(envFile, keyFile string, oldEnv, oldKey []byte, hadEnv, hadKey bool) {
	if hadEnv {
		_ = os.WriteFile(envFile, oldEnv, 0600)
		_ = core.ChmodPrivateFile(envFile)
	} else {
		_ = os.Remove(envFile)
	}
	if hadKey {
		_ = os.WriteFile(keyFile, oldKey, 0600)
		_ = core.ChmodPrivateFile(keyFile)
	} else {
		_ = os.Remove(keyFile)
	}
}

func copyPrivateFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0600); err != nil {
		return err
	}
	return core.ChmodPrivateFile(dst)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
