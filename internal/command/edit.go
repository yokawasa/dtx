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

	tmpDir, err := os.MkdirTemp("", "dtx-edit-*")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpEnvFile := core.SameBasePath(tmpDir, env)
	targetEnvFile := ctx.Paths.EnvFile(env)
	targetKeyFile := ctx.Paths.KeyFile(env)
	isNewEnv := !ctx.Store.EnvExists(env)

	if !isNewEnv {
		if !ctx.Store.KeyExists(env) {
			return fmt.Errorf("key for env %q does not exist", env)
		}
		plain, err := ctx.Dotenvx.Decrypt(targetEnvFile, targetKeyFile, verbose)
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

	if err := ctx.Dotenvx.Encrypt(tmpEnvFile, targetKeyFile, verbose); err != nil {
		return fmt.Errorf("failed to encrypt env %q: %w", env, err)
	}
	if isNewEnv && !ctx.Store.KeyExists(env) {
		return fmt.Errorf("env %q does not contain variables to encrypt", env)
	}

	if err := os.MkdirAll(filepath.Dir(targetEnvFile), 0700); err != nil {
		return fmt.Errorf("create env directory: %w", err)
	}
	if err := os.Rename(tmpEnvFile, targetEnvFile); err != nil {
		return fmt.Errorf("save env file: %w", err)
	}
	if err := core.ChmodPrivateFile(targetEnvFile); err != nil {
		return err
	}
	if ctx.Store.KeyExists(env) {
		if err := core.ChmodPrivateFile(targetKeyFile); err != nil {
			return err
		}
	}

	fmt.Fprintf(ctx.Stdout, "Edited env: %s\n", env)
	return nil
}
