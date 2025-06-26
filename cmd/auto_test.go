// cmd/auto_test.go

package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bgreenwell/gitego/config"
)

func TestAutoCommand(t *testing.T) {
	// 1. Setup
	mockCfg := &config.Config{
		Profiles: map[string]*config.Profile{
			"work": {Name: "Work User", Email: "work@example.com"},
		},
		AutoRules: []*config.AutoRule{},
	}

	var savedConfig bool
	var ensuredProfile, includedProfile string
	var includedPath string

	// 2. Create the test runner with mocked dependencies
	runner := &autoRunner{
		load: func() (*config.Config, error) {
			cfgCopy := *mockCfg
			return &cfgCopy, nil
		},
		save: func(c *config.Config) error {
			savedConfig = true
			mockCfg = c
			return nil
		},
		ensureProfileGitconfig: func(profileName string, p *config.Profile) error {
			ensuredProfile = profileName
			return nil
		},
		addIncludeIf: func(profileName, path string) error {
			includedProfile = profileName
			includedPath = path
			return nil
		},
	}

	// 3. Execute the command's logic
	// Use a path that is simple and cross-platform for the test.
	testPath := filepath.Join("tmp", "work")
	args := []string{testPath, "work"}
	runner.run(autoCmd, args)

	// 4. Assertions
	if !savedConfig {
		t.Error("Expected the config to be saved, but it wasn't.")
	}

	if len(mockCfg.AutoRules) != 1 {
		t.Fatalf("Expected 1 auto-rule to be added, but found %d", len(mockCfg.AutoRules))
	}

	rule := mockCfg.AutoRules[0]
	if rule.Profile != "work" {
		t.Errorf("Expected rule to be for profile 'work', got '%s'", rule.Profile)
	}

	// Check that the path stored in the rule is absolute and has forward slashes
	absTestPath, _ := filepath.Abs(testPath)
	expectedPath := filepath.ToSlash(absTestPath) + "/"
	if rule.Path != expectedPath {
		t.Errorf("Expected rule path to be '%s', got '%s'", expectedPath, rule.Path)
	}

	if ensuredProfile != "work" {
		t.Error("Expected EnsureProfileGitconfig to be called for 'work' profile.")
	}

	if includedProfile != "work" {
		t.Error("Expected AddIncludeIf to be called for 'work' profile.")
	}

	if !strings.HasSuffix(includedPath, "/") {
		t.Errorf("Expected path passed to AddIncludeIf to have a trailing slash, got '%s'", includedPath)
	}
}
