package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/yokawasa/dtx/internal/apperr"
	"github.com/yokawasa/dtx/internal/command"
)

const usage = `Usage:
  dtx use <env>
  dtx current
  dtx ls
  dtx run [env] [--verbose] -- <command>
  dtx edit <env> [--verbose]
  dtx completion <bash|zsh|fish>

Commands:
  dtx use <env>                           Set the current env
  dtx current                             Print the current env
  dtx ls                                  List available envs
  dtx run [env] [--verbose] -- <command>  Run a command with an env
  dtx edit <env> [--verbose]              Create or edit an encrypted env
  dtx completion <bash|zsh|fish>          Generate shell completion
`

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		_, _ = fmt.Fprint(stderr, usage)
		return apperr.Silent(2)
	}
	if args[0] == "-h" || args[0] == "--help" {
		_, _ = fmt.Fprint(stdout, usage)
		return nil
	}
	if args[0] == "-v" || args[0] == "--version" {
		_, _ = fmt.Fprintln(stdout, "dtx "+Version)
		return nil
	}

	ctx, err := command.NewContext(stdin, stdout, stderr)
	if err != nil {
		return err
	}

	switch args[0] {
	case "use":
		if len(args) != 2 {
			return printUsageError(stderr, "usage: dtx use <env>")
		}
		return command.Use(ctx, args[1])
	case "current":
		if len(args) != 1 {
			return printUsageError(stderr, "usage: dtx current")
		}
		return command.Current(ctx)
	case "ls":
		if len(args) != 1 {
			return printUsageError(stderr, "usage: dtx ls")
		}
		return command.List(ctx)
	case "run":
		opts, err := parseRunArgs(args[1:])
		if err != nil {
			if isUsageError(err) {
				return printUsageError(stderr, err.Error())
			}
			return err
		}
		return command.Run(ctx, opts)
	case "edit":
		env, verbose, err := parseEditArgs(args[1:])
		if err != nil {
			if isUsageError(err) {
				return printUsageError(stderr, err.Error())
			}
			return err
		}
		return command.Edit(ctx, env, verbose)
	case "completion":
		shell, err := parseCompletionArgs(args[1:])
		if err != nil {
			if isUsageError(err) {
				return printUsageError(stderr, err.Error())
			}
			return err
		}
		_, err = io.WriteString(stdout, completionScript(shell))
		return err
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func isUsageError(err error) bool {
	var usageErr usageError
	return errors.As(err, &usageErr)
}

func printUsageError(stderr io.Writer, message string) error {
	_, _ = fmt.Fprintln(stderr, message)
	return apperr.Silent(2)
}
