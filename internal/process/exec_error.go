package process

import (
	"errors"
	"os/exec"
)

func AsExitError(err error, target **exec.ExitError) bool {
	return errors.As(err, target)
}
