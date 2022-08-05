// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"

	"github.com/spf13/afero"

	"github.com/ava-labs/apm/storage"
)

type UpgradeConfig struct {
	Executor Executor

	RepoFactory storage.RepositoryFactory
	StateFile   storage.StateFile

	TmpPath    string
	PluginPath string
	Installer  Installer
	Fs         afero.Fs
}

func NewUpgrade(config UpgradeConfig) *Upgrade {
	return &Upgrade{
		executor:    config.Executor,
		repoFactory: config.RepoFactory,
		tmpPath:     config.TmpPath,
		pluginPath:  config.PluginPath,
		installer:   config.Installer,
		stateFile:   config.StateFile,
		fs:          config.Fs,
	}
}

type Upgrade struct {
	executor Executor

	repoFactory storage.RepositoryFactory
	stateFile   storage.StateFile

	tmpPath    string
	pluginPath string

	installer Installer
	fs        afero.Fs
}

func (u *Upgrade) Execute() error {
	upgraded := false

	for name := range u.stateFile.InstalledVMs {
		wf := NewUpgradeVM(UpgradeVMConfig{
			Executor:    u.executor,
			RepoFactory: u.repoFactory,
			FullVMName:  name,
			StateFile:   u.stateFile,
			TmpPath:     u.tmpPath,
			PluginPath:  u.pluginPath,
			Installer:   u.installer,
			Fs:          u.fs,
		})

		err := u.executor.Execute(wf)
		if err == nil || err == ErrAlreadyUpdated {
			upgraded = true
		} else if err != nil {
			return err
		}
	}

	if !upgraded {
		fmt.Printf("No changes detected.\n")
		return nil
	}

	return nil
}
