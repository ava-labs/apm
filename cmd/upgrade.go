// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func upgrade(fs afero.Fs) *cobra.Command {
	// this flag is optional
	vm := ""
	command := &cobra.Command{
		Use: "upgrade",
		Short: "Upgrades a virtual machine. If none is specified, all " +
			"installed virtual machines are upgraded.",
	}
	command.PersistentFlags().StringVar(&vm, "vm", "", "vm alias to install")
	command.RunE = func(_ *cobra.Command, _ []string) error {
		apm, err := initAPM(fs)
		if err != nil {
			return err
		}

		return apm.Upgrade(vm)
	}

	return command
}
