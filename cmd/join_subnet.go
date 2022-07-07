// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func joinSubnet(fs afero.Fs) *cobra.Command {
	subnet := ""

	command := &cobra.Command{
		Use:   "join-subnet",
		Short: "Installs all virtual machines for a subnet.",
	}

	command.PersistentFlags().StringVar(&subnet, "subnet", "", "subnet alias to join")
	err := command.MarkPersistentFlagRequired("subnet")
	if err != nil {
		panic(err)
	}

	command.RunE = func(_ *cobra.Command, _ []string) error {
		apm, err := initAPM(fs)
		if err != nil {
			return err
		}

		return apm.JoinSubnet(subnet)
	}

	return command
}
