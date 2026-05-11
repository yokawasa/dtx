package dotenvx

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/yokawasa/dtx/internal/process"
)

type Adapter struct {
	Binary string
	Runner process.Runner
	Stderr io.Writer
}

func NewAdapter(stdin io.Reader, stdout io.Writer, stderr io.Writer) Adapter {
	return Adapter{
		Binary: "dotenvx",
		Runner: process.Runner{
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: stderr,
		},
		Stderr: stderr,
	}
}

func (a Adapter) CheckAvailable() error {
	if _, err := exec.LookPath(a.Binary); err != nil {
		return fmt.Errorf("dotenvx is required but not found in PATH (install dotenvx: https://dotenvx.com/docs/install)")
	}
	return nil
}

func (a Adapter) Run(envFile string, keyFile string, command []string, verbose bool) (int, error) {
	if err := a.CheckAvailable(); err != nil {
		return 1, err
	}
	if len(command) == 0 {
		return 1, fmt.Errorf("command is required")
	}

	args := a.baseArgs(verbose, "run")
	args = append(args, "--strict", "-f", envFile, "-fk", keyFile, "--")
	args = append(args, command...)
	code, err := a.Runner.Run(a.Binary, args, nil)
	if err != nil {
		return code, fmt.Errorf("run dotenvx: %w", err)
	}
	return code, nil
}

func (a Adapter) Decrypt(envFile string, keyFile string, verbose bool) ([]byte, error) {
	if err := a.CheckAvailable(); err != nil {
		return nil, err
	}

	args := a.baseArgs(verbose, "decrypt")
	args = append(args, "--stdout", "-f", envFile, "-fk", keyFile)
	stdout, stderr, code, err := a.Runner.Capture(a.Binary, args, nil)
	if err != nil {
		return nil, fmt.Errorf("run dotenvx decrypt: %w", err)
	}
	if code != 0 {
		a.writeCapturedStderr(stderr)
		return nil, fmt.Errorf("failed to decrypt env file (exit %d)", code)
	}
	if verbose {
		a.writeCapturedStderr(stderr)
	}
	return stdout, nil
}

func (a Adapter) Encrypt(envFile string, keyFile string, verbose bool) error {
	if err := a.CheckAvailable(); err != nil {
		return err
	}

	args := a.baseArgs(verbose, "encrypt")
	args = append(args, "-f", envFile, "-fk", keyFile)
	if verbose {
		code, err := a.Runner.Run(a.Binary, args, nil)
		if err != nil {
			return fmt.Errorf("run dotenvx encrypt: %w", err)
		}
		if code != 0 {
			return fmt.Errorf("failed to encrypt env file (exit %d)", code)
		}
		return nil
	}

	stdout, stderr, code, err := a.Runner.Capture(a.Binary, args, nil)
	_ = stdout
	if err != nil {
		return fmt.Errorf("run dotenvx encrypt: %w", err)
	}
	if code != 0 {
		a.writeCapturedStderr(stderr)
		return fmt.Errorf("failed to encrypt env file (exit %d)", code)
	}
	return nil
}

func (a Adapter) baseArgs(verbose bool, command string) []string {
	if verbose {
		return []string{"--verbose", command}
	}
	return []string{"--quiet", command}
}

func (a Adapter) writeCapturedStderr(stderr []byte) {
	if len(stderr) > 0 && a.Stderr != nil {
		_, _ = a.Stderr.Write(stderr)
	}
}
