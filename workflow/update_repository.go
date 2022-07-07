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
)

var (
	subnetDir = "subnets"
	vmDir     = "vms"

	subnetKey = "subnet"
	vmKey     = "vm"

	_ Workflow = &UpdateRepository{}
)

type UpdateRepositoryConfig struct {
	RepoName       string
	RepositoryPath string

	AliasBytes []byte

	PreviousCommit plumbing.Hash
	LatestCommit   plumbing.Hash

	SourceInfo  storage.SourceInfo
	Repository  storage.Repository
	Registry    storage.Storage[storage.RepoList]
	SourcesList storage.Storage[storage.SourceInfo]

	Fs afero.Fs
}

func NewUpdateRepository(config UpdateRepositoryConfig) *UpdateRepository {
	return &UpdateRepository{
		repoName:           config.RepoName,
		repositoryPath:     config.RepositoryPath,
		aliasBytes:         config.AliasBytes,
		previousCommit:     config.PreviousCommit,
		latestCommit:       config.LatestCommit,
		repository:         config.Repository,
		registry:           config.Registry,
		sourcesList:        config.SourcesList,
		repositoryMetadata: config.SourceInfo,
		fs:                 config.Fs,
	}
}

type UpdateRepository struct {
	repoName       string
	repositoryPath string

	aliasBytes []byte

	previousCommit plumbing.Hash
	latestCommit   plumbing.Hash

	repository  storage.Repository
	registry    storage.Storage[storage.RepoList]
	sourcesList storage.Storage[storage.SourceInfo]

	repositoryMetadata storage.SourceInfo

	fs afero.Fs
}

func (u *UpdateRepository) Execute() error {
	if err := u.update(); err != nil {
		fmt.Printf("Unexpected error while updating definitions. %s", err)
		return err
	}

	// checkpoint progress
	updatedCheckpoint := storage.SourceInfo{
		Alias:  u.repositoryMetadata.Alias,
		URL:    u.repositoryMetadata.URL,
		Commit: u.latestCommit,
	}
	if err := u.sourcesList.Put(u.aliasBytes, updatedCheckpoint); err != nil {
		return err
	}

	fmt.Printf("Finished update.\n")

	return nil
}

func (u *UpdateRepository) update() error {
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
