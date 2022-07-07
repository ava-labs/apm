// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func install(fs afero.Fs) *cobra.Command {
	vm := ""
	command := &cobra.Command{
		Use:   "install-vm",
		Short: "Installs a virtual machine by its alias",
	}
	command.PersistentFlags().StringVar(&vm, "vm", "", "vm alias to install")
	err := command.MarkPersistentFlagRequired("vm")
	if err != nil {
		panic(err)
	}

	command.RunE = func(_ *cobra.Command, _ []string) error {
		apm, err := initAPM(fs)
		if err != nil {
			return err
		}

		return apm.Install(vm)
	}

	return command
}
