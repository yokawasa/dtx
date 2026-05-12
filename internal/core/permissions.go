package core

import (
	"fmt"
	"os"
)

func ChmodPrivateFile(path string) error {
	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("chmod %s: %w", path, err)
	}
	return nil
}
