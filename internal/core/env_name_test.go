package core

import "testing"

func TestValidateEnvName(t *testing.T) {
	valid := []string{"dev", "prod-1", "staging_us", "qa.v2", "A1"}
	for _, env := range valid {
		if err := ValidateEnvName(env); err != nil {
			t.Fatalf("ValidateEnvName(%q) returned error: %v", env, err)
		}
	}

	invalid := []string{"", "../prod", "prod/key", "-prod", ".prod", ".."}
	for _, env := range invalid {
		if err := ValidateEnvName(env); err == nil {
			t.Fatalf("ValidateEnvName(%q) returned nil", env)
		}
	}
}
