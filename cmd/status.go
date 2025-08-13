// cmd/status.go

package cmd

import (
	"github.com/bgreenwell/gitego/config"
	"github.com/bgreenwell/gitego/utils"
	"github.com/spf13/cobra"
)

// statusRunner holds dependencies for the status command for mocking.
type statusRunner struct {
	load         func() (*config.Config, error)
	getGitConfig func(string) (string, error)
}

// run contains the core logic for the status command.
func (sr *statusRunner) run(cmd *cobra.Command, args []string) {
	name, errName := sr.getGitConfig("user.name")
	email, errEmail := sr.getGitConfig("user.email")

	if errName != nil || errEmail != nil {
		cmd.PrintErrln("Not inside a Git repository or user not configured.")

		return
	}

	cfg, err := sr.load()
	if err != nil {
		cmd.PrintErrf("Warning: Could not load gitego config: %v\n", err)
	}

	source := "Global Git Config"

	if cfg != nil {
		// This will check the current directory against the loaded rules.
		_, ruleSource := cfg.GetActiveProfileForCurrentDir()
		if ruleSource != "No active gitego profile" {
			source = ruleSource
		}
	}

	cmd.Println("--- Git Identity Status ---")
	cmd.Printf("  Name:   %s\n", name)
	cmd.Printf("  Email:  %s\n", email)
	cmd.Printf("  Source: %s\n", source)
	cmd.Println("---------------------------")
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Displays the current effective Git user and any active gitego rule.",
	Long: `Checks the current Git configuration and any applicable gitego rules
to show you which user.name and user.email are currently in effect. It also
tells you whether the configuration is coming from your global .gitconfig or
from a gitego auto-switch rule.`,
	Run: func(cmd *cobra.Command, args []string) {
		runner := &statusRunner{
			load:         config.Load,
			getGitConfig: utils.GetEffectiveGitConfig,
		}
		runner.run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
