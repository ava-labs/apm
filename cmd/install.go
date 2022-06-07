package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ava-labs/apm/apm"
)

func install() *cobra.Command {
	vmAlias := ""
	command := &cobra.Command{
		Use:   "install-vm",
		Short: "installs a virtual machine by its alias",
	}
	command.PersistentFlags().StringVar(&vmAlias, "vm-alias", "", "vm alias to install")
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
		return apm.Install(vmAlias)
	}

	return command
}
