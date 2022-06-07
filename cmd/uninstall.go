package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ava-labs/apm/apm"
)

func uninstall(apm *apm.APM) *cobra.Command {
	vmAlias := ""
	command := &cobra.Command{
		Use:   "uninstall-vm",
		Short: "uninstalls a virtual machine by its alias",
	}
	command.PersistentFlags().StringVar(&vmAlias, "vm-alias", "", "vm alias to install")
	command.RunE = func(_ *cobra.Command, _ []string) error {
		return apm.Uninstall(vmAlias)
	}

	return command
}
