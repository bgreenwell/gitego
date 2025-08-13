// cmd/rm_test.go

package cmd

import (
	"testing"

	"github.com/bgreenwell/gitego/config"
)

// setupRmTestConfig creates a mock config for rm command testing.
func setupRmTestConfig() *config.Config {
	return &config.Config{
		Profiles: map[string]*config.Profile{
			"work":     {Name: "Work User", Email: "work@example.com"},
			"personal": {Name: "Personal User", Email: "personal@example.com"},
		},
		AutoRules: []*config.AutoRule{
			{Path: "/path/to/work", Profile: "work"},
			{Path: "/path/to/personal", Profile: "personal"},
		},
	}
}

func TestRmCommand(t *testing.T) {
	// Setup: Create mock config and state trackers
	mockCfg := setupRmTestConfig()

	var removedIncludeIf, removedProfileCfg, deletedToken string

	var saved bool

	// Create a test runner with mock functions
	runner := &rmRunner{
		load: func() (*config.Config, error) {
			cfgCopy := *mockCfg

			return &cfgCopy, nil
		},
		save: func(c *config.Config) error {
			saved = true
			*mockCfg = *c

			return nil
		},
		removeIncludeIf: func(profileName string) error {
			removedIncludeIf = profileName

			return nil
		},
		removeProfileCfg: func(profileName string) error {
			removedProfileCfg = profileName

			return nil
		},
		deleteToken: func(profileName string) error {
			deletedToken = profileName

			return nil
		},
	}

	// Execute the command to remove the "work" profile
	args := []string{"work"}
	forceFlag = true; runner.run(rmCmd, args)

	forceFlag = false

	// Assertions
	validateProfileRemoval(t, mockCfg)
	validateRmCommandEffects(t, saved, removedIncludeIf, removedProfileCfg, deletedToken)
}

// validateProfileRemoval validates that the profile was properly removed.
func validateProfileRemoval(t *testing.T, mockCfg *config.Config) {
	t.Helper()

	if _, exists := mockCfg.Profiles["work"]; exists {
		t.Error("Expected 'work' profile to be deleted from config, but it still exists.")
	}

	if len(mockCfg.Profiles) != 1 {
		t.Errorf("Expected 1 profile to remain, but found %d", len(mockCfg.Profiles))
	}

	if len(mockCfg.AutoRules) != 1 || mockCfg.AutoRules[0].Profile != "personal" {
		t.Error("Expected auto-rule for 'work' profile to be removed.")
	}
}

// validateRmCommandEffects validates all side effects of the rm command.
func validateRmCommandEffects(t *testing.T, saved bool, removedIncludeIf, removedProfileCfg, deletedToken string) {
	t.Helper()

	if !saved {
		t.Error("Expected config.Save() to be called, but it wasn't.")
	}

	if removedIncludeIf != "work" {
		t.Error("Expected RemoveIncludeIf to be called for 'work' profile.")
	}

	if removedProfileCfg != "work" {
		t.Error("Expected the profile config file for 'work' to be removed.")
	}

	if deletedToken != "work" {
		t.Error("Expected DeleteToken to be called for 'work' profile.")
	}
}

