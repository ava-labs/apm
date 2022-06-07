package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ava-labs/apm/apm"
)

func joinSubnet(apm *apm.APM) *cobra.Command {
	subnetAlias := ""

	command := &cobra.Command{
		Use:   "join-subnet",
		Short: "join a subnet by its alias.",
	}

	command.PersistentFlags().StringVar(&subnetAlias, "subnet-alias", "", "subnet alias to join")
	command.RunE = func(_ *cobra.Command, _ []string) error {
		return apm.JoinSubnet(subnetAlias)
	}

	return command
}
