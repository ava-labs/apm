package workflow

import (
	"fmt"

	"github.com/spf13/afero"

	"github.com/ava-labs/apm/storage"
)

type UpgradeConfig struct {
	Executor Executor

	RepoFactory  storage.RepositoryFactory
	Registry     storage.Storage[storage.RepoList]
	SourcesList  storage.Storage[storage.SourceInfo]
	InstalledVMs storage.Storage[storage.InstallInfo]

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

	repoFactory storage.RepositoryFactory
	registry    storage.Storage[storage.RepoList]

	installedVMs storage.Storage[storage.InstallInfo]
	sourcesList  storage.Storage[storage.SourceInfo]

	tmpPath    string
	pluginPath string

	installer Installer
	fs        afero.Fs
}

func (u *Upgrade) Execute() error {
	upgraded := false

	itr := u.installedVMs.Iterator()
	defer itr.Release()

	for itr.Next() {
		wf := NewUpgradeVM(UpgradeVMConfig{
			Executor:     u.executor,
			RepoFactory:  u.repoFactory,
			FullVMName:   string(itr.Key()),
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
