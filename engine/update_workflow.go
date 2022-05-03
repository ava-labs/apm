package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanchego/database"
	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/repository"
	"github.com/ava-labs/apm/types"
)

var (
	subnetDir = "subnets"
	vmDir     = "vms"

	subnetKey = "subnet"
	vmKey     = "vm"

	_ Workflow = &UpdateWorkflow{}
)

type UpdateWorkflowConfig struct {
	RepoName       string
	RepositoryPath string

	AliasBytes []byte

	PreviousCommit plumbing.Hash
	LatestCommit   plumbing.Hash

	RepoRegistry   repository.Group
	GlobalRegistry repository.Group

	RepositoryMetadata repository.Metadata

	RepositoryDB database.Database
}

func NewUpdateWorkflow(config UpdateWorkflowConfig) *UpdateWorkflow {
	return &UpdateWorkflow{
		repoName:           config.RepoName,
		repositoryPath:     config.RepositoryPath,
		aliasBytes:         config.AliasBytes,
		previousCommit:     config.PreviousCommit,
		latestCommit:       config.LatestCommit,
		repoRegistry:       config.RepoRegistry,
		globalRegistry:     config.GlobalRegistry,
		repositoryMetadata: config.RepositoryMetadata,
		repositoryDB:       config.RepositoryDB,
	}
}

type UpdateWorkflow struct {
	repoName       string
	repositoryPath string

	aliasBytes []byte

	previousCommit plumbing.Hash
	latestCommit   plumbing.Hash

	repoRegistry   repository.Group
	globalRegistry repository.Group

	repositoryMetadata repository.Metadata

	repositoryDB database.Database
}

func (u UpdateWorkflow) Execute() error {
	repoVMs := u.repoRegistry.VMs()
	repoSubnets := u.repoRegistry.Subnets()
	vmsPath := filepath.Join(u.repositoryPath, vmDir)

	if err := loadFromYAML[*types.VM](vmKey, vmsPath, u.aliasBytes, u.latestCommit, u.globalRegistry.VMs(), repoVMs); err != nil {
		return err
	}

	subnetsPath := filepath.Join(u.repositoryPath, subnetDir)
	if err := loadFromYAML[*types.Subnet](subnetKey, subnetsPath, u.aliasBytes, u.latestCommit, u.globalRegistry.Subnets(), repoSubnets); err != nil {
		return err
	}

	// Now we need to delete anything that wasn't updated in the latest commit
	if err := deleteStalePlugins[*types.VM](repoVMs, u.latestCommit); err != nil {
		return err
	}
	if err := deleteStalePlugins[*types.Subnet](repoSubnets, u.latestCommit); err != nil {
		return err
	}

	updatedMetadata := repository.Metadata{
		Alias:  u.repositoryMetadata.Alias,
		URL:    u.repositoryMetadata.URL,
		Commit: u.latestCommit,
	}
	updatedMetadataBytes, err := yaml.Marshal(updatedMetadata)
	if err != nil {
		return err
	}

	if err := u.repositoryDB.Put(u.aliasBytes, updatedMetadataBytes); err != nil {
		return err
	}

	if u.previousCommit == plumbing.ZeroHash {
		fmt.Printf("Finished initializing %s@%s.\n", u.repoName, u.latestCommit)
	} else {
		fmt.Printf("Finished updating from %s to %s@%s.\n", u.previousCommit, u.repoName, u.latestCommit)
	}

	return nil
}

func loadFromYAML[T types.Plugin](
	key string,
	path string,
	repositoryAlias []byte,
	commit plumbing.Hash,
	globalDB database.Database,
	repoDB database.Database,
) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	globalBatch := globalDB.NewBatch()
	repoBatch := repoDB.NewBatch()

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

		fileBytes, err := os.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return err
		}
		data := make(map[string]T)

		if err := yaml.Unmarshal(fileBytes, data); err != nil {
			return err
		}
		record := &repository.Plugin[T]{
			Plugin: data[key],
			Commit: commit,
		}

		alias := data[key].Alias()
		aliasBytes := []byte(alias)

		registry := &repository.Registry{}
		registryBytes, err := globalDB.Get(aliasBytes)
		if err == database.ErrNotFound {
			registry = &repository.Registry{
				Repositories: []string{},
			}
		} else if err != nil {
			return err
		} else if err := yaml.Unmarshal(registryBytes, registry); err != nil {
			return err
		}

		registry.Repositories = append(registry.Repositories, string(repositoryAlias))

		updatedRegistryBytes, err := yaml.Marshal(registry)
		if err != nil {
			return err
		}
		updatedRecordBytes, err := yaml.Marshal(record)
		if err != nil {
			return err
		}

		if err := globalBatch.Put(aliasBytes, updatedRegistryBytes); err != nil {
			return err
		}
		if err := repoBatch.Put(aliasBytes, updatedRecordBytes); err != nil {
			return err
		}

		fmt.Printf("Updated plugin definition in registry for %s:%s@%s.\n", repositoryAlias, alias, commit)
	}

	if err := globalBatch.Write(); err != nil {
		return err
	}
	if err := repoBatch.Write(); err != nil {
		return err
	}

	return nil
}

func deleteStalePlugins[T types.Plugin](db database.Database, latestCommit plumbing.Hash) error {
	itr := db.NewIterator()
	batch := db.NewBatch()

	for itr.Next() {
		record := &repository.Plugin[T]{}
		if err := yaml.Unmarshal(itr.Value(), record); err != nil {
			return nil
		}

		if record.Commit != latestCommit {
			fmt.Printf("Deleting a stale plugin: %s@%s as of %s.\n", record.Plugin.Alias(), record.Commit, latestCommit)
			if err := batch.Delete(itr.Key()); err != nil {
				return err
			}
		}
	}

	if err := batch.Write(); err != nil {
		return err
	}
	return nil
}
