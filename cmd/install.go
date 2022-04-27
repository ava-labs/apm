package cmd

import (
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
			Directory:        viper.GetString(ApmPathKey),
			Auth:             credentials,
			AdminApiEndpoint: viper.GetString(AdminApiEndpoint),
			PluginDir:        viper.GetString(PluginPathKey),
		})
		if err != nil {
			return err
		}
		return apm.Install(vmAlias)
	}

	return command
}
