package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ava-labs/apm/apm"
)

func update(apm *apm.APM) *cobra.Command {
	command := &cobra.Command{
		Use:   "update",
		Short: "updates all registries and virtual machines on the node",
	}
	command.RunE = func(_ *cobra.Command, _ []string) error {
		return apm.Update()
	}

	return command
}
