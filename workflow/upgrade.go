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

	RepoFactory  storage.RepositoryFactory
	Registry     map[string]storage.RepoList
	SourcesList  map[string]storage.SourceInfo
	InstalledVMs map[string]storage.InstallInfo

	TmpPath    string
	PluginPath string
	Installer  Installer
	Fs         afero.Fs
}

func NewUpgrade(config UpgradeConfig) *Upgrade {
	return &Upgrade{
		executor:     config.Executor,
		repoFactory:  config.RepoFactory,
		registry:     config.Registry,
		installedVMs: config.InstalledVMs,
		tmpPath:      config.TmpPath,
		pluginPath:   config.PluginPath,
		installer:    config.Installer,
		sourcesList:  config.SourcesList,
		fs:           config.Fs,
	}
}

type Upgrade struct {
	executor Executor

	repoFactory  storage.RepositoryFactory
	registry     map[string]storage.RepoList
	sourcesList  map[string]storage.SourceInfo
	installedVMs map[string]storage.InstallInfo

	tmpPath    string
	pluginPath string

	installer Installer
	fs        afero.Fs
}

func (u *Upgrade) Execute() error {
	upgraded := false

	for name := range u.installedVMs {
		wf := NewUpgradeVM(UpgradeVMConfig{
			Executor:     u.executor,
			RepoFactory:  u.repoFactory,
			FullVMName:   name,
			InstalledVMs: u.installedVMs,
			TmpPath:      u.tmpPath,
			PluginPath:   u.pluginPath,
			Installer:    u.installer,
			Fs:           u.fs,
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
