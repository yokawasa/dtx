package cli

import (
	"fmt"

	"github.com/yokawasa/dtx/internal/command"
)

func parseRunArgs(args []string) (command.RunOptions, error) {
	opts := command.RunOptions{}
	separator := -1
	for i, arg := range args {
		if arg == "--" {
			separator = i
			break
		}
	}
	if separator == -1 {
		return opts, newUsageError("usage: dtx run [env] [--verbose] -- <command>")
	}

	prefix := args[:separator]
	for _, arg := range prefix {
		switch arg {
		case "--verbose":
			opts.Verbose = true
		default:
			if opts.Env != "" {
				return opts, newUsageError("usage: dtx run [env] [--verbose] -- <command>")
			}
			opts.Env = arg
		}
	}

	opts.Command = args[separator+1:]
	if len(opts.Command) == 0 {
		return opts, fmt.Errorf("command is required after --")
	}
	return opts, nil
}

func parseEditArgs(args []string) (string, bool, error) {
	var env string
	var verbose bool
	for _, arg := range args {
		switch arg {
		case "--verbose":
			verbose = true
		default:
			if env != "" {
				return "", false, newUsageError("usage: dtx edit <env> [--verbose]")
			}
			env = arg
		}
	}
	if env == "" {
		return "", false, newUsageError("usage: dtx edit <env> [--verbose]")
	}
	return env, verbose, nil
}

func parseCompletionArgs(args []string) (string, error) {
	if len(args) != 1 {
		return "", newUsageError("usage: dtx completion <bash|zsh|fish>")
	}

	switch args[0] {
	case "bash", "zsh", "fish":
		return args[0], nil
	default:
		return "", newUsageError("usage: dtx completion <bash|zsh|fish>")
	}
}

type usageError string

func (e usageError) Error() string {
	return string(e)
}

func newUsageError(message string) error {
	return usageError(message)
}
