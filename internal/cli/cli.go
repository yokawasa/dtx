package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/yokawasa/dtx/internal/command"
)

const usage = `Usage:
  dtx use <env>
  dtx current
  dtx ls
  dtx run [env] [--verbose] -- <command>
  dtx edit <env> [--verbose]
`

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf(strings.TrimSpace(usage))
	}
	if args[0] == "-h" || args[0] == "--help" {
		_, _ = fmt.Fprint(stdout, usage)
		return nil
	}
	if args[0] == "-v" || args[0] == "--version" {
		_, _ = fmt.Fprintln(stdout, "dtx dev")
		return nil
	}

	ctx, err := command.NewContext(stdin, stdout, stderr)
	if err != nil {
		return err
	}

	switch args[0] {
	case "use":
		if len(args) != 2 {
			return fmt.Errorf("usage: dtx use <env>")
		}
		return command.Use(ctx, args[1])
	case "current":
		if len(args) != 1 {
			return fmt.Errorf("usage: dtx current")
		}
		return command.Current(ctx)
	case "ls":
		if len(args) != 1 {
			return fmt.Errorf("usage: dtx ls")
		}
		return command.List(ctx)
	case "run":
		opts, err := parseRunArgs(args[1:])
		if err != nil {
			return err
		}
		return command.Run(ctx, opts)
	case "edit":
		env, verbose, err := parseEditArgs(args[1:])
		if err != nil {
			return err
		}
		return command.Edit(ctx, env, verbose)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
