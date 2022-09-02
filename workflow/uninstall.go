// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/ava-labs/apm/state"
)

var _ Workflow = &Uninstall{}

func NewUninstall(config UninstallConfig) *Uninstall {
	return &Uninstall{
		name:       config.Name,
		repoAlias:  config.RepoAlias,
		plugin:     config.Plugin,
		stateFile:  config.StateFile,
		fs:         config.Fs,
		pluginPath: config.PluginPath,
	}
}

type UninstallConfig struct {
	Name       string
	Plugin     string
	RepoAlias  string
	StateFile  state.File
	Fs         afero.Fs
	PluginPath string
}

type Uninstall struct {
	name       string
	plugin     string
	repoAlias  string
	stateFile  state.File
	fs         afero.Fs
	pluginPath string
}

func (u Uninstall) Execute() error {
	installInfo, ok := u.stateFile.InstallationRegistry[u.name]
	if !ok {
		fmt.Printf("VM %s is already not installed. Skipping.\n", u.name)
		return nil
	}

	vmPath := filepath.Join(u.pluginPath, installInfo.ID)

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

	delete(u.stateFile.InstallationRegistry, u.name)
	fmt.Printf("Successfully uninstalled %s.\n", u.name)

	return nil
}
