package command

import (
	"fmt"
)

func Use(ctx Context, env string) error {
	if err := validateEnvArg(env); err != nil {
		return err
	}
	if err := ctx.Paths.Ensure(); err != nil {
		return err
	}
	if !ctx.Store.EnvExists(env) {
		return fmt.Errorf("env %q does not exist", env)
	}
	if err := ctx.Store.WriteCurrent(env); err != nil {
		return err
	}
	fmt.Fprintf(ctx.Stdout, "Using env: %s\n", env)
	return nil
}
