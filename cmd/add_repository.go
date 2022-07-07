// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func addRepository(fs afero.Fs) *cobra.Command {
	url := ""
	alias := ""
	command := &cobra.Command{
		Use:   "add-repository",
		Short: "Adds a custom repository to the list of tracked repositories",
	}
	command.PersistentFlags().StringVar(&alias, "alias", "", "alias for the repository")
	command.MarkPersistentFlagRequired("alias")

	command.PersistentFlags().StringVar(&url, "url", "", "url to the repository")
	command.MarkPersistentFlagRequired("url")

	command.RunE = func(_ *cobra.Command, _ []string) error {
		apm, err := initAPM(fs)
		if err != nil {
			return err
		}

		return apm.AddRepository(alias, url)
	}

	return command
}
