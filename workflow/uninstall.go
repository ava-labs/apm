package workflow

import (
	"fmt"

	"github.com/ava-labs/avalanchego/version"

	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/types"
)

var _ Workflow = &Uninstall{}

func NewUninstall(config UninstallConfig) *Uninstall {
	return &Uninstall{
		name:         config.Name,
		repoAlias:    config.RepoAlias,
		plugin:       config.Plugin,
		vmStorage:    config.VMStorage,
		installedVMs: config.InstalledVMs,
	}
}

type UninstallConfig struct {
	Name         string
	Plugin       string
	RepoAlias    string
	VMStorage    storage.Storage[storage.Definition[types.VM]]
	InstalledVMs storage.Storage[version.Semantic]
}

type Uninstall struct {
	name         string
	plugin       string
	repoAlias    string
	vmStorage    storage.Storage[storage.Definition[types.VM]]
	installedVMs storage.Storage[version.Semantic]
}

func (u Uninstall) Execute() error {
	ok, err := u.installedVMs.Has([]byte(u.name))
	if err != nil {
		return err
	}

	if !ok {
		fmt.Printf("VM %s is already not installed. Skipping.\n", u.name)
		return nil
	}

	ok, err = u.vmStorage.Has([]byte(u.plugin))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf(fmt.Sprintf("Virtual machine %s doesn't exist under the repository for %s.", u.plugin, u.repoAlias))
	}

	if err := u.installedVMs.Delete([]byte(u.name)); err != nil {
		return err
	}

	fmt.Printf("Successfully uninstalled %s.\n", u.name)

	return nil
}
