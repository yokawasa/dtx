package core

import (
	"fmt"
	"regexp"
)

var envNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

func ValidateEnvName(env string) error {
	if env == "" {
		return fmt.Errorf("env name is required")
	}
	if len(env) > 128 {
		return fmt.Errorf("invalid env name %q: must be 128 characters or fewer", env)
	}
	if env == "." || env == ".." || !envNamePattern.MatchString(env) {
		return fmt.Errorf("invalid env name %q: use letters, numbers, dot, underscore, or hyphen", env)
	}
	return nil
}
