// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/ava-labs/avalanchego/utils/perms"
	"github.com/spf13/afero"

	"github.com/ava-labs/apm/checksum"
	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/types"
)

var _ Workflow = &Install{}

type InstallConfig struct {
	Name         string
	Plugin       string
	Organization string
	Repo         string
	TmpPath      string
	PluginPath   string

	InstalledVMs storage.Storage[storage.InstallInfo]
	VMStorage    storage.Storage[storage.Definition[types.VM]]
	Fs           afero.Fs
	Installer    Installer
}

func NewInstall(config InstallConfig) *Install {
	return &Install{
		name:         config.Name,
		plugin:       config.Plugin,
		organization: config.Organization,
		repo:         config.Repo,
		tmpPath:      config.TmpPath,
		pluginPath:   config.PluginPath,
		installedVMs: config.InstalledVMs,
		vmStorage:    config.VMStorage,
		fs:           config.Fs,
		installer:    config.Installer,
		checksummer:  checksum.NewSHA256(config.Fs),
	}
}

type Install struct {
	name         string
	plugin       string
	organization string
	repo         string
	tmpPath      string
	pluginPath   string

	installedVMs storage.Storage[storage.InstallInfo]
	vmStorage    storage.Storage[storage.Definition[types.VM]]
	fs           afero.Fs
	installer    Installer
	checksummer  checksum.Checksummer
}

func (i Install) Execute() error {
	var (
		definition storage.Definition[types.VM]
		err        error
	)

	definition, err = i.vmStorage.Get([]byte(i.plugin))
	if err != nil {
		return err
	}

	vm := definition.Definition

	archiveFile := fmt.Sprintf("%s.tar.gz", i.plugin)
	tmpPath := filepath.Join(i.tmpPath, i.organization, i.repo)
	archiveFilePath := filepath.Join(tmpPath, archiveFile)
	workingDir := filepath.Join(tmpPath, i.plugin)

	if err := i.installer.Download(vm.URL, archiveFilePath); err != nil {
		// TODO sometimes these aren't cleaned up if we fail before cleanup step
		return err
	}

	fmt.Printf("Calculating checksums...\n")
	hash := fmt.Sprintf("%x", i.checksummer.Checksum(archiveFilePath))
	if hash != vm.SHA256 {
		return fmt.Errorf("checksums did not match. Expected %s but saw %s", vm.SHA256, hash)
	}

	fmt.Printf("Saw expected checksum value of %s\n", hash)

	// Create the directory we'll store the plugin sources in if it doesn't exist.
	if _, err := i.fs.Stat(workingDir); errors.Is(err, fs.ErrNotExist) {
		fmt.Printf("Creating sources directory...\n")
		if err := i.fs.Mkdir(workingDir, perms.ReadWriteExecute); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	fmt.Printf("Unpacking %s...\n", i.name)
	if err := i.installer.Decompress(archiveFilePath, workingDir); err != nil {
		return err
	}

	if vm.InstallScript != "" {
		args := strings.Split(vm.InstallScript, " ")
		fmt.Printf("Running install script at %s...\n", vm.InstallScript)
		if err := i.installer.Install(workingDir, args...); err != nil {
			return err
		}
	} else {
		fmt.Printf("No install script found for %s.\n", i.name)
	}

	fmt.Printf("Moving binary %s into plugin directory...\n", vm.ID)
	if err := i.fs.Rename(filepath.Join(workingDir, vm.BinaryPath), filepath.Join(i.pluginPath, vm.ID)); err != nil {
		return err
	}

	fmt.Printf("Cleaning up temporary files...\n")
	if err := i.fs.Remove(filepath.Join(tmpPath, archiveFile)); err != nil {
		return err
	}

	if err := i.fs.RemoveAll(filepath.Join(tmpPath, i.plugin)); err != nil {
		return err
	}

	fmt.Printf("Adding virtual machine %s to installation registry...\n", vm.ID)
	installInfo := storage.InstallInfo{
		ID:      vm.ID,
		Version: vm.Version,
	}
	if err := i.installedVMs.Put([]byte(i.name), installInfo); err != nil {
		return err
	}

	fmt.Printf("Successfully installed %s@v%v.%v.%v in %s\n", i.name, vm.Version.Major, vm.Version.Minor, vm.Version.Patch, filepath.Join(i.pluginPath, vm.ID))
	return nil
}
