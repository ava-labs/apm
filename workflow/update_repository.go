// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/ava-labs/avalanchego/database"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/types"
	"github.com/ava-labs/apm/util"
)

var (
	subnetDir = "subnets"
	vmDir     = "vms"

	subnetKey = "subnet"
	vmKey     = "vm"

	_ Workflow = &UpdateRepository{}
)

type UpdateRepositoryConfig struct {
	Executor       Executor
	RepoName       string
	RepositoryPath string

	AliasBytes []byte

	PreviousCommit plumbing.Hash
	LatestCommit   plumbing.Hash

	SourceInfo   storage.SourceInfo
	Repository   storage.Repository
	Registry     storage.Storage[storage.RepoList]
	SourcesList  storage.Storage[storage.SourceInfo]
	InstalledVMs storage.Storage[storage.InstallInfo]

	TmpPath    string
	PluginPath string
	Installer  Installer
	Fs         afero.Fs
}

func NewUpdateRepository(config UpdateRepositoryConfig) *UpdateRepository {
	return &UpdateRepository{
		executor:           config.Executor,
		repoName:           config.RepoName,
		repositoryPath:     config.RepositoryPath,
		aliasBytes:         config.AliasBytes,
		previousCommit:     config.PreviousCommit,
		latestCommit:       config.LatestCommit,
		repository:         config.Repository,
		registry:           config.Registry,
		repositoryMetadata: config.SourceInfo,
		installedVMs:       config.InstalledVMs,
		sourcesList:        config.SourcesList,
		tmpPath:            config.TmpPath,
		pluginPath:         config.PluginPath,
		installer:          config.Installer,
		fs:                 config.Fs,
	}
}

type UpdateRepository struct {
	executor       Executor
	repoName       string
	repositoryPath string

	aliasBytes []byte

	previousCommit plumbing.Hash
	latestCommit   plumbing.Hash

	repository storage.Repository
	registry   storage.Storage[storage.RepoList]

	repositoryMetadata storage.SourceInfo

	installedVMs storage.Storage[storage.InstallInfo]
	sourcesList  storage.Storage[storage.SourceInfo]

	tmpPath    string
	pluginPath string

	installer Installer
	fs        afero.Fs
}

func (u *UpdateRepository) Execute() error {
	if err := u.updateDefinitions(); err != nil {
		fmt.Printf("Unexpected error while updating definitions. %s", err)
		return err
	}

	if err := u.updateVMs(); err != nil {
		fmt.Printf("Unexpected error while updating vms. %s", err)
		return err
	}

	// checkpoint progress
	updatedMetadata := storage.SourceInfo{
		Alias:  u.repositoryMetadata.Alias,
		URL:    u.repositoryMetadata.URL,
		Commit: u.latestCommit,
	}
	if err := u.sourcesList.Put(u.aliasBytes, updatedMetadata); err != nil {
		return err
	}

	fmt.Printf("Finished update.\n")

	return nil
}

func (u *UpdateRepository) updateDefinitions() error {
	vmsPath := filepath.Join(u.repositoryPath, vmDir)

	if err := loadFromYAML[types.VM](u.fs, vmKey, vmsPath, u.aliasBytes, u.latestCommit, u.registry, u.repository.VMs); err != nil {
		return err
	}

	subnetsPath := filepath.Join(u.repositoryPath, subnetDir)
	if err := loadFromYAML[types.Subnet](u.fs, subnetKey, subnetsPath, u.aliasBytes, u.latestCommit, u.registry, u.repository.Subnets); err != nil {
		return err
	}

	// Now we need to delete anything that wasn't updated in the latest commit
	if err := deleteStaleDefinitions[types.VM](u.repository.VMs, u.latestCommit); err != nil {
		return err
	}
	if err := deleteStaleDefinitions[types.Subnet](u.repository.Subnets, u.latestCommit); err != nil {
		return err
	}

	if u.previousCommit == plumbing.ZeroHash {
		fmt.Printf("Finished initializing definitions for%s@%s.\n", u.repoName, u.latestCommit)
	} else {
		fmt.Printf("Finished updating definitions from %s to %s@%s.\n", u.previousCommit, u.repoName, u.latestCommit)
	}

	return nil
}

func loadFromYAML[T types.Definition](
	fs afero.Fs,
	key string,
	path string,
	repositoryAlias []byte,
	commit plumbing.Hash,
	registry storage.Storage[storage.RepoList],
	repository storage.Storage[storage.Definition[T]],
) error {
	files, err := afero.ReadDir(fs, path)
	if err != nil {
		return err
	}
	// globalBatch := registry.NewBatch()
	// repoBatch := repository.NewBatch()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		nameWithExtension := file.Name()
		// Strip any extension from the file. This is to support windows .exe
		// files.
		name := nameWithExtension[:len(nameWithExtension)-len(filepath.Ext(nameWithExtension))]

		// Skip hidden files.
		if len(name) == 0 {
			continue
		}

		fileBytes, err := afero.ReadFile(fs, filepath.Join(path, file.Name()))
		if err != nil {
			return err
		}
		data := make(map[string]T)

		if err := yaml.Unmarshal(fileBytes, data); err != nil {
			return err
		}
		definition := storage.Definition[T]{
			Definition: data[key],
			Commit:     commit,
		}

		alias := data[key].GetAlias()
		aliasBytes := []byte(alias)

		repoList, err := registry.Get(aliasBytes)
		if err == database.ErrNotFound {
			repoList = storage.RepoList{ // TODO check if this can be removed
				Repositories: []string{},
			}
		} else if err != nil {
			return err
		}

		repositoryAliasStr := string(repositoryAlias)
		idx := sort.SearchStrings(repoList.Repositories, repositoryAliasStr)

		if idx == len(repoList.Repositories) {
			repoList.Repositories = append(repoList.Repositories, repositoryAliasStr)
		} else if repoList.Repositories[idx] != repositoryAliasStr {
			repoList.Repositories = append(repoList.Repositories[:idx+1], repoList.Repositories[idx:]...)
			repoList.Repositories[idx] = repositoryAliasStr
		}

		if err := registry.Put(aliasBytes, repoList); err != nil {
			return err
		}
		if err := repository.Put(aliasBytes, definition); err != nil {
			return err
		}

		fmt.Printf("Updated plugin definition in registry for %s:%s@%s.\n", repositoryAlias, alias, commit)
	}

	return nil
}

func deleteStaleDefinitions[T types.Definition](db storage.Storage[storage.Definition[T]], latestCommit plumbing.Hash) error {
	itr := db.Iterator()
	defer itr.Release()
	// TODO batching

	for itr.Next() {
		definition, err := itr.Value()
		if err != nil {
			return err
		}

		if definition.Commit != latestCommit {
			fmt.Printf("Deleting a stale plugin: %s@%s as of %s.\n", definition.Definition.GetAlias(), definition.Commit, latestCommit)
			if err := db.Delete(itr.Key()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (u *UpdateRepository) updateVMs() error {
	updated := false

	itr := u.installedVMs.Iterator()
	defer itr.Release()

	for itr.Next() {
		fullVMName := string(itr.Key())
		installInfo, err := itr.Value()
		if err != nil {
			return err
		}

		repoAlias, vmName := util.ParseQualifiedName(fullVMName)
		organization, repo := util.ParseAlias(repoAlias)

		var definition storage.Definition[types.VM]

		vmStorage := u.repository.VMs
		definition, err = vmStorage.Get([]byte(vmName))
		if err == database.ErrNotFound {
			fmt.Printf("Warning - found a vm while updating %s which is no longer registered in a repository. You should uninstall this VM to avoid noisy logs. Skipping...\n", fullVMName)
			continue
		}
		if err != nil {
			return err
		}

		updatedVM := definition.Definition

		if installInfo.Version.Compare(&updatedVM.Version) < 0 {
			fmt.Printf(
				"Detected an update for %s from v%v.%v.%v to v%v.%v.%v.\n",
				fullVMName,
				installInfo.Version.Major,
				installInfo.Version.Minor,
				installInfo.Version.Patch,
				updatedVM.Version.Major,
				updatedVM.Version.Minor,
				updatedVM.Version.Patch,
			)
			installWorkflow := NewInstall(InstallConfig{
				Name:         fullVMName,
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
				fullVMName,
				updatedVM.Version.Major,
				updatedVM.Version.Minor,
				updatedVM.Version.Patch,
			)
			if err := u.executor.Execute(installWorkflow); err != nil {
				return err
			}

			updated = true
		}
	}

	if !updated {
		fmt.Printf("No changes detected.\n")
		return nil
	}

	return nil
}
