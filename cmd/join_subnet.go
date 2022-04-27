package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ava-labs/apm/apm"
)

func joinSubnet() *cobra.Command {
	subnetAlias := ""

	command := &cobra.Command{
		Use:   "join",
		Short: "join a subnet by its alias.",
	}

	command.PersistentFlags().StringVar(&subnetAlias, "subnet-alias", "", "subnet alias to join")
	command.RunE = func(_ *cobra.Command, _ []string) error {
		credentials, err := getCredentials()
		if err != nil {
			return err
		}
		apm, err := apm.New(apm.Config{
			Directory: apmDir,
			Auth:      credentials,
		})

		if err != nil {
			return err
		}

		return apm.JoinSubnet(subnetAlias)
	}

	return command
}
