package workflow

import (
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/database"
	"github.com/spf13/afero"

	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/types"
	"github.com/ava-labs/apm/util"
)

var ErrAlreadyUpdated = errors.New("already up-to-date")

type UpgradeVMConfig struct {
	Executor Executor

	FullVMName   string
	InstalledVMs storage.Storage[storage.InstallInfo]

	TmpPath    string
	PluginPath string
	Installer  Installer
	Fs         afero.Fs
}

func NewUpgradeVM(config UpgradeVMConfig) *UpgradeVM {
	return &UpgradeVM{
		executor:     config.Executor,
		fullVMName:   config.FullVMName,
		installedVMs: config.InstalledVMs,
		tmpPath:      config.TmpPath,
		pluginPath:   config.PluginPath,
		installer:    config.Installer,
		fs:           config.Fs,
	}
}

type UpgradeVM struct {
	fullVMName string
	executor   Executor

	aliasBytes []byte

	repository storage.Repository

	installedVMs storage.Storage[storage.InstallInfo]

	tmpPath    string
	pluginPath string

	installer Installer
	fs        afero.Fs
}

func (u *UpgradeVM) Execute() error {
	installInfo, err := u.installedVMs.Get([]byte(u.fullVMName))
	if err != nil {
		return err
	}

	repoAlias, vmName := util.ParseQualifiedName(u.fullVMName)
	organization, repo := util.ParseAlias(repoAlias)

	var definition storage.Definition[types.VM]

	vmStorage := u.repository.VMs
	definition, err = vmStorage.Get([]byte(vmName))
	if err == database.ErrNotFound {
		fmt.Printf("Warning - found a vm while upgrading %s which is no longer registered in a repository. You should uninstall this VM to avoid noisy logs. Skipping...\n", u.fullVMName)
		return nil
	}
	if err != nil {
		return err
	}

	upgradedVM := definition.Definition

	if installInfo.Version.Compare(&upgradedVM.Version) < 0 {
		fmt.Printf(
			"Detected an upgrade for %s from v%v.%v.%v to v%v.%v.%v.\n",
			u.fullVMName,
			installInfo.Version.Major,
			installInfo.Version.Minor,
			installInfo.Version.Patch,
			upgradedVM.Version.Major,
			upgradedVM.Version.Minor,
			upgradedVM.Version.Patch,
		)
		installWorkflow := NewInstall(InstallConfig{
			Name:         u.fullVMName,
			Plugin:       vmName,
			Organization: organization,
			Repo:         repo,
			TmpPath:      u.tmpPath,
			PluginPath:   u.pluginPath,
			InstalledVMs: u.installedVMs,
			VMStorage:    u.repository.VMs,
			Installer:    u.installer,
		})

		fmt.Printf(
			"Rebuilding binaries for %s v%v.%v.%v.\n",
			u.fullVMName,
			upgradedVM.Version.Major,
			upgradedVM.Version.Minor,
			upgradedVM.Version.Patch,
		)
		if err := u.executor.Execute(installWorkflow); err != nil {
			return err
		}
	}

	return ErrAlreadyUpdated
}
