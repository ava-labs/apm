package workflow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanchego/utils/perms"
	"github.com/ava-labs/avalanchego/version"
	"github.com/spf13/afero"

	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/types"
)

var _ Workflow = &InstallWorkflow{}

type InstallWorkflowConfig struct {
	Name         string
	Plugin       string
	Organization string
	Repo         string
	TmpPath      string
	PluginPath   string

	InstalledVMs storage.Storage[version.Semantic]
	VMStorage    storage.Storage[storage.Definition[types.VM]]
	Fs           afero.Fs
	Installer    Installer
}

func NewInstallWorkflow(config InstallWorkflowConfig) *InstallWorkflow {
	return &InstallWorkflow{
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
	}
}

type InstallWorkflow struct {
	name         string
	plugin       string
	organization string
	repo         string
	tmpPath      string
	pluginPath   string

	installedVMs storage.Storage[version.Semantic]
	vmStorage    storage.Storage[storage.Definition[types.VM]]
	fs           afero.Fs
	installer    Installer
}

func (i InstallWorkflow) Execute() error {
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
	// Create the directory we'll store the plugin sources in if it doesn't exist.
	if _, err := i.fs.Stat(i.plugin); errors.Is(err, os.ErrNotExist) {
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
		fmt.Printf("Running install script at %s...\n", vm.InstallScript)
		if err := i.installer.Install(workingDir, vm.InstallScript); err != nil {
			return err
		}
	} else {
		fmt.Printf("No install script found for %s.", i.name)
	}

	fmt.Printf("Moving binary %s into plugin directory...\n", vm.ID_)
	if err := i.fs.Rename(filepath.Join(workingDir, vm.BinaryPath), filepath.Join(i.pluginPath, vm.ID_)); err != nil {
		return err
	}

	fmt.Printf("Cleaning up temporary files...\n")
	if err := i.fs.Remove(filepath.Join(tmpPath, archiveFile)); err != nil {
		return err
	}

	if err := i.fs.RemoveAll(filepath.Join(tmpPath, i.plugin)); err != nil {
		return err
	}

	fmt.Printf("Adding virtual machine %s to installation registry...\n", vm.ID_)
	if err := i.installedVMs.Put([]byte(i.name), vm.Version); err != nil {
		return err
	}

	fmt.Printf("Successfully installed %s@v%v.\n", i.name, vm.Version.Str)
	return nil
}
