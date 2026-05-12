package command

import (
	"fmt"

	"github.com/yokawasa/dtx/internal/apperr"
)

type RunOptions struct {
	Env     string
	Verbose bool
	Command []string
}

func Run(ctx Context, opts RunOptions) error {
	env := opts.Env
	if env == "" {
		current, err := ctx.Store.ReadCurrent()
		if err != nil {
			return err
		}
		env = current
	}
	if err := validateEnvArg(env); err != nil {
		return err
	}
	if err := ctx.Store.RequireRunnableEnv(env); err != nil {
		return err
	}
	if len(opts.Command) == 0 {
		return fmt.Errorf("command is required after --")
	}

	fmt.Fprintf(ctx.Stdout, "Using env: %s\n", env)
	code, err := ctx.Dotenvx.Run(ctx.Paths.EnvFile(env), ctx.Paths.KeyFile(env), opts.Command, opts.Verbose)
	if err != nil {
		return err
	}
	if code != 0 {
		return apperr.Silent(code)
	}
	return nil
}
