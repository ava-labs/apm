package engine

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/utils/perms"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/repository"
	"github.com/ava-labs/apm/types"
	"github.com/ava-labs/apm/url"
)

var _ Workflow = &InstallWorkflow{}

type InstallWorkflowConfig struct {
	Name         string
	Plugin       string
	Organization string
	Repo         string
	TmpPath      string
	PluginPath   string

	InstalledVMs database.Database
	Registry     repository.Registry
	HttpClient   url.Client
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
		registry:     config.Registry,
		httpClient:   config.HttpClient,
	}
}

type InstallWorkflow struct {
	name         string
	plugin       string
	organization string
	repo         string
	tmpPath      string
	pluginPath   string

	installedVMs database.Database
	registry     repository.Registry
	httpClient   url.Client
}

func (i InstallWorkflow) Execute() error {
	bytes, err := i.registry.VMs().Get([]byte(i.plugin))
	if err != nil {
		return err
	}

	definition := repository.Definition[types.VM]{}
	if err := yaml.Unmarshal(bytes, &definition); err != nil {
		return err
	}

	vm := definition.Definition
	archiveFile := fmt.Sprintf("%s.tar.gz", i.plugin)
	tmpPath := filepath.Join(i.tmpPath, i.organization, i.repo)

	if vm.InstallScript == "" {
		fmt.Printf("No install script found for %s.", i.name)
		return nil
	}

	// Download the .tar.gz file from the url
	if err := i.httpClient.Download(filepath.Join(tmpPath, archiveFile), vm.URL); err != nil {
		return err
	}

	// Create the directory we'll store the plugin sources in if it doesn't exist.
	if _, err := os.Stat(i.plugin); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Creating sources directory...\n")
		if err := os.Mkdir(filepath.Join(tmpPath, i.plugin), perms.ReadWriteExecute); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	fmt.Printf("Unpacking %s...\n", i.name)
	cmd := exec.Command("tar", "xf", archiveFile, "-C", i.plugin, "--strip-components", "1")
	cmd.Dir = tmpPath
	if err := cmd.Run(); err != nil {
		return err
	}

	workingDir := filepath.Join(tmpPath, i.plugin)
	fmt.Printf("Running install script at %s...\n", vm.InstallScript)
	cmd = exec.Command(vm.InstallScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = workingDir
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Printf("Moving binary %s into plugin directory...\n", vm.ID_)
	if err := os.Rename(filepath.Join(workingDir, vm.BinaryPath), filepath.Join(i.pluginPath, vm.ID_)); err != nil {
		return err
	}

	fmt.Printf("Cleaning up temporary files...\n")
	if err := os.Remove(filepath.Join(tmpPath, archiveFile)); err != nil {
		return err
	}

	if err := os.RemoveAll(filepath.Join(tmpPath, i.plugin)); err != nil {
		return err
	}

	fmt.Printf("Adding virtual machine %s to installation registry...\n", vm.ID_)
	installedVersion, err := yaml.Marshal(vm.Version)
	if err != nil {
		return err
	}
	if err := i.installedVMs.Put([]byte(i.name), installedVersion); err != nil {
		return err
	}

	fmt.Printf("Successfully installed %s@v%v.\n", i.name, vm.Version.Str)
	return nil
}
