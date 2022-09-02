// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"

	"github.com/spf13/afero"

	"github.com/ava-labs/apm/git"
	"github.com/ava-labs/apm/state"
)

type UpgradeConfig struct {
	Executor Executor

	RepoFactory state.RepositoryFactory
	StateFile   state.File

	TmpPath    string
	PluginPath string
	Installer  Installer
	Git        git.Factory
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
		git:         config.Git,
		fs:          config.Fs,
	}
}

type Upgrade struct {
	executor Executor

	repoFactory state.RepositoryFactory
	stateFile   state.File

	tmpPath    string
	pluginPath string

	installer Installer
	git       git.Factory
	fs        afero.Fs
}

func (u *Upgrade) Execute() error {
	upgraded := false

	for name := range u.stateFile.InstallationRegistry {
		wf := NewUpgradeVM(UpgradeVMConfig{
			Executor:    u.executor,
			FullVMName:  name,
			RepoFactory: u.repoFactory,
			StateFile:   u.stateFile,
			TmpPath:     u.tmpPath,
			PluginPath:  u.pluginPath,
			Installer:   u.installer,
			Git:         u.git,
			Fs:          u.fs,
		})

		if err := u.executor.Execute(wf); err == ErrAlreadyUpdated {
			continue
		} else if err != nil {
			return err
		}

		upgraded = true
	}

	if !upgraded {
		fmt.Printf("No changes detected.\n")
		return nil
	}

	return nil
}
