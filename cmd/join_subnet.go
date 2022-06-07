package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ava-labs/apm/apm"
)

func joinSubnet() *cobra.Command {
	subnetAlias := ""

	command := &cobra.Command{
		Use:   "join-subnet",
		Short: "join a subnet by its alias.",
	}

	command.PersistentFlags().StringVar(&subnetAlias, "subnet-alias", "", "subnet alias to join")
	command.RunE = func(_ *cobra.Command, _ []string) error {
		credentials, err := getCredentials()
		if err != nil {
			return err
		}
		apm, err := apm.New(apm.Config{
			Directory:        viper.GetString(apmPathKey),
			Auth:             credentials,
			AdminApiEndpoint: viper.GetString(adminApiEndpointKey),
			PluginDir:        viper.GetString(pluginPathKey),
			Fs:               afero.NewOsFs(),
		})

		if err != nil {
			return err
		}

		return apm.JoinSubnet(subnetAlias)
	}

	return command
}
