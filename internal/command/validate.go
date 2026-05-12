package command

import "github.com/yokawasa/dtx/internal/core"

func validateEnvArg(env string) error {
	return core.ValidateEnvName(env)
}
