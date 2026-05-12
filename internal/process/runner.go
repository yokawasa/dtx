package process

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Runner struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (r Runner) Run(name string, args []string, env []string) (int, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = r.Stdin
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	if env != nil {
		cmd.Env = env
	}

	err := cmd.Run()
	if err == nil {
		return 0, nil
	}

	var exitErr *exec.ExitError
	if ok := AsExitError(err, &exitErr); ok {
		return exitErr.ExitCode(), nil
	}
	return 1, err
}

func (r Runner) Capture(name string, args []string, env []string) ([]byte, []byte, int, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = r.Stdin
	if env != nil {
		cmd.Env = env
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	outData := stdout.Bytes()
	errData := stderr.Bytes()
	if err == nil {
		return outData, errData, 0, nil
	}

	var exitErr *exec.ExitError
	if ok := AsExitError(err, &exitErr); ok {
		return outData, errData, exitErr.ExitCode(), nil
	}
	return outData, errData, 1, err
}

func (r Runner) RunEditor(editor, path string) error {
	fields := strings.Fields(editor)
	if len(fields) == 0 {
		return fmt.Errorf("editor is not set")
	}

	args := append(fields[1:], path)
	cmd := exec.Command(fields[0], args...)
	cmd.Stdin = r.Stdin
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	if cmd.Stdin == nil {
		cmd.Stdin = os.Stdin
	}
	if cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}
