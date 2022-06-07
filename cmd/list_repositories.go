package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ava-labs/apm/apm"
)

func listRepositories(apm *apm.APM) *cobra.Command {
	command := &cobra.Command{
		Use:   "list-repositories",
		Short: "list registered source repositories for avalanche plugins",
	}
	command.RunE = func(_ *cobra.Command, _ []string) error {
		return apm.ListRepositories()
	}

	return command
}
