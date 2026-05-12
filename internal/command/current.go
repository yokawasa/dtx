package command

import "fmt"

func Current(ctx Context) error {
	env, err := ctx.Store.ReadCurrent()
	if err != nil {
		return err
	}
	fmt.Fprintln(ctx.Stdout, env)
	return nil
}
