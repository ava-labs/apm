package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ava-labs/apm/apm"
)

func update() *cobra.Command {
	command := &cobra.Command{
		Use:   "update",
		Short: "updates all registries and virtual machines on the node",
	}
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
		})
		if err != nil {
			return err
		}
		return apm.Update()
	}

	return command
}
