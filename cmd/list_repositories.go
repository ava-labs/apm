// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func listRepositories(fs afero.Fs) *cobra.Command {
	command := &cobra.Command{
		Use:   "list-repositories",
		Short: "Lists all tracked plugin repositories.",
	}
	command.RunE = func(_ *cobra.Command, _ []string) error {
		apm, err := initAPM(fs)
		if err != nil {
			return err
		}

		return apm.ListRepositories()
	}

	return command
}
