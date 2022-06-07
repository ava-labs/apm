package workflow

import (
	"fmt"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/version"

	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/util"
)

var _ Workflow = &Uninstall{}

func NewUninstall(config UninstallConfig) *Uninstall {
	return &Uninstall{
		name:         config.Name,
		db:           config.DB,
		installedVMs: config.InstalledVMs,
	}
}

type UninstallConfig struct {
	Name         string
	DB           database.Database
	InstalledVMs storage.Storage[version.Semantic]
}

type Uninstall struct {
	name         string
	db           database.Database
	installedVMs storage.Storage[version.Semantic]
}

func (u Uninstall) Execute() error {
	nameBytes := []byte(u.name)

	ok, err := u.installedVMs.Has(nameBytes)
	if err != nil {
		return err
	}

	if !ok {
		fmt.Printf("VM %s is already not installed. Skipping.\n", u.name)
		return nil
	}

	alias, plugin := util.ParseQualifiedName(u.name)

	repoDB := prefixdb.New([]byte(alias), u.db)
	repoVMDB := prefixdb.New(vmPrefix, repoDB)

	ok, err = repoVMDB.Has([]byte(plugin))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Printf("Virtual machine already %s doesn't exist in the vm registry for %s.", u.name, alias)
		return nil
	}

	if err := u.installedVMs.Delete([]byte(plugin)); err != nil {
		return err
	}

	fmt.Printf("Successfully uninstalled %s.", u.name)

	return nil
}
