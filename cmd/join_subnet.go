package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func joinSubnet(fs afero.Fs) *cobra.Command {
	subnetAlias := ""

	command := &cobra.Command{
		Use:   "join-subnet",
		Short: "join a subnet by its alias.",
	}

	command.PersistentFlags().StringVar(&subnetAlias, "subnet-alias", "", "subnet alias to join")
	command.RunE = func(_ *cobra.Command, _ []string) error {
		apm, err := initAPM(fs)
		if err != nil {
			return err
		}

		return apm.JoinSubnet(subnetAlias)
	}

	return command
}
