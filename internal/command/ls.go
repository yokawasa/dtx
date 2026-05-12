package command

import "fmt"

func List(ctx Context) error {
	envs, err := ctx.Store.List()
	if err != nil {
		return err
	}
	for _, env := range envs {
		fmt.Fprintln(ctx.Stdout, env)
	}
	return nil
}
