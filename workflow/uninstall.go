// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/ava-labs/avalanchego/database"
	"github.com/spf13/afero"

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
		fs:           config.Fs,
		pluginPath:   config.PluginPath,
	}
}

type UninstallConfig struct {
	Name         string
	Plugin       string
	RepoAlias    string
	VMStorage    storage.Storage[storage.Definition[types.VM]]
	InstalledVMs storage.Storage[storage.InstallInfo]
	Fs           afero.Fs
	PluginPath   string
}

type Uninstall struct {
	name         string
	plugin       string
	repoAlias    string
	vmStorage    storage.Storage[storage.Definition[types.VM]]
	installedVMs storage.Storage[storage.InstallInfo]
	fs           afero.Fs
	pluginPath   string
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

	vm, err := u.vmStorage.Get([]byte(u.plugin))
	if err == database.ErrNotFound {
		// If we don't have the definition, provide a warning log. It's possible
		// this used to exist and was removed for whatever reason. In that case,
		// we should still remove it from our installation registry to unblock
		// the user.
		fmt.Printf("Virtual machine %s doesn't exist under the repository for %s. Continuing uninstall anyways...\n", u.plugin, u.repoAlias)
	} else if err != nil {
		return err
	}

	vmPath := filepath.Join(u.pluginPath, vm.Definition.GetID())

	switch _, err := u.fs.Stat(vmPath); err {
	case nil:
		fmt.Printf("Deleting %s...\n", vmPath)
		if err := u.fs.Remove(vmPath); err != nil {
			return err
		}
	default:
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Printf("%s doesn't exist already. Nothing to delete here.\n", vmPath)
		} else {
			return err
		}
	}

	if err := u.installedVMs.Delete([]byte(u.name)); err != nil {
		return err
	}
	fmt.Printf("Successfully uninstalled %s.\n", u.name)

	return nil
}
