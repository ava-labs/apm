package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func update(fs afero.Fs) *cobra.Command {
	command := &cobra.Command{
		Use:   "update",
		Short: "updates all registries and virtual machines on the node",
	}
	command.RunE = func(_ *cobra.Command, _ []string) error {
		apm, err := initAPM(fs)
		if err != nil {
			return err
		}

		return apm.Update()
	}

	return command
}
