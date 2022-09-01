// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/afero"

	"github.com/ava-labs/apm/git"
	"github.com/ava-labs/apm/state"
	"github.com/ava-labs/apm/util"
)

var ErrAlreadyUpdated = errors.New("already up-to-date")

type UpgradeVMConfig struct {
	Executor Executor

	FullVMName  string
	RepoFactory state.RepositoryFactory
	StateFile   state.File

	TmpPath    string
	PluginPath string
	Installer  Installer
	Fs         afero.Fs
	Git        git.Factory
}

func NewUpgradeVM(config UpgradeVMConfig) *UpgradeVM {
	return &UpgradeVM{
		executor:    config.Executor,
		fullVMName:  config.FullVMName,
		repoFactory: config.RepoFactory,
		stateFile:   config.StateFile,
		tmpPath:     config.TmpPath,
		pluginPath:  config.PluginPath,
		installer:   config.Installer,
		fs:          config.Fs,
		git:         config.Git,
	}
}

type UpgradeVM struct {
	fullVMName string
	executor   Executor

	repoFactory state.RepositoryFactory

	stateFile state.File

	tmpPath    string
	pluginPath string

	installer Installer
	fs        afero.Fs
	git       git.Factory
}

func (u *UpgradeVM) Execute() error {
	installInfo := u.stateFile.InstallationRegistry[u.fullVMName]

	repoAlias, vmName := util.ParseQualifiedName(u.fullVMName)
	organization, repo := util.ParseAlias(repoAlias)

	repository, err := u.repoFactory.GetRepository(repoAlias)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Warning - found a repository %s while upgrading %s "+
			"which is no longer downloaded. You might need to re-add this "+
			"repository and call update, or uninstall this vm to avoid noisy logs. "+
			"Skipping...\n", repoAlias, u.fullVMName)
		return nil
	} else if err != nil {
		return err
	}

	if _, err := repository.GetVM(vmName); err != nil {
		fmt.Printf("Warning - found a vm while upgrading %s which is no "+
			"longer registered in a repository. You should uninstall this VM to "+
			"avoid noisy logs. Skipping...\n", u.fullVMName)
		return nil
	}

	latest, err := u.git.GetLastModified(repository.GetPath(), fmt.Sprintf("vms/%s.%s", vmName, "yaml"))
	if err != nil {
		return err
	}

	if installInfo.Commit == latest {
		return ErrAlreadyUpdated
	}

	fmt.Printf(
		"Detected an upgrade for %s from %s to %s\n",
		u.fullVMName,
		installInfo.Commit,
		latest,
	)
	wf := NewInstall(InstallConfig{
		Name:         u.fullVMName,
		Plugin:       vmName,
		Organization: organization,
		Repo:         repo,
		TmpPath:      u.tmpPath,
		PluginPath:   u.pluginPath,
		StateFile:    u.stateFile,
		Repository:   repository,
		Installer:    u.installer,
		Fs:           u.fs,
	})

	fmt.Printf(
		"Rebuilding binaries for %s@%s\n",
		u.fullVMName,
		latest,
	)
	return u.executor.Execute(wf)
}
