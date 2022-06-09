package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func install(fs afero.Fs) *cobra.Command {
	vmAlias := ""
	command := &cobra.Command{
		Use:   "install-vm",
		Short: "installs a virtual machine by its alias",
	}
	command.PersistentFlags().StringVar(&vmAlias, "vm-alias", "", "vm alias to install")
	command.RunE = func(_ *cobra.Command, _ []string) error {
		apm, err := initAPM(fs)
		if err != nil {
			return err
		}

		return apm.Install(vmAlias)
	}

	return command
}
