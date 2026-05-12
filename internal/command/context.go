package command

import (
	"io"

	"github.com/yokawasa/dtx/internal/core"
	"github.com/yokawasa/dtx/internal/dotenvx"
	"github.com/yokawasa/dtx/internal/process"
)

type Context struct {
	Paths   core.Paths
	Store   core.EnvStore
	Dotenvx dotenvx.Adapter
	Runner  process.Runner
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

func NewContext(stdin io.Reader, stdout io.Writer, stderr io.Writer) (Context, error) {
	paths, err := core.NewPaths()
	if err != nil {
		return Context{}, err
	}

	return Context{
		Paths:   paths,
		Store:   core.NewEnvStore(paths),
		Dotenvx: dotenvx.NewAdapter(stdin, stdout, stderr),
		Runner: process.Runner{
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: stderr,
		},
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}, nil
}
