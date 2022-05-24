package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/version"
	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/repository"
	"github.com/ava-labs/apm/types"
	"github.com/ava-labs/apm/url"
	"github.com/ava-labs/apm/util"
)

var (
	subnetDir = "subnets"
	vmDir     = "vms"

	subnetKey = "subnet"
	vmKey     = "vm"

	_ Workflow = &UpdateWorkflow{}
)

type UpdateWorkflowConfig struct {
	Engine         Engine
	RepoName       string
	RepositoryPath string

	AliasBytes []byte

	PreviousCommit plumbing.Hash
	LatestCommit   plumbing.Hash

	RepoRegistry   repository.Registry
	GlobalRegistry repository.Registry

	RepositoryMetadata repository.Metadata

	RepositoryDB database.Database
	InstalledVMs database.Database
	DB           database.Database

	TmpPath    string
	PluginPath string
	HttpClient url.Client
}

func NewUpdateWorkflow(config UpdateWorkflowConfig) *UpdateWorkflow {
	return &UpdateWorkflow{
		engine:             config.Engine,
		repoName:           config.RepoName,
		repositoryPath:     config.RepositoryPath,
		aliasBytes:         config.AliasBytes,
		previousCommit:     config.PreviousCommit,
		latestCommit:       config.LatestCommit,
		repoRegistry:       config.RepoRegistry,
		globalRegistry:     config.GlobalRegistry,
		repositoryMetadata: config.RepositoryMetadata,
		installedVMs:       config.InstalledVMs,
		repositoryDB:       config.RepositoryDB,
		db:                 config.DB,
		tmpPath:            config.TmpPath,
		pluginPath:         config.PluginPath,
		httpClient:         config.HttpClient,
	}
}

type UpdateWorkflow struct {
	engine         Engine
	repoName       string
	repositoryPath string

	aliasBytes []byte

	previousCommit plumbing.Hash
	latestCommit   plumbing.Hash

	repoRegistry   repository.Registry
	globalRegistry repository.Registry

	repositoryMetadata repository.Metadata

	installedVMs database.Database
	repositoryDB database.Database
	db           database.Database

	tmpPath    string
	pluginPath string

	httpClient url.Client
}

func (u *UpdateWorkflow) updateVMs() error {
	updated := false

	itr := u.installedVMs.NewIterator()

	for itr.Next() {
		fullVMName := string(itr.Key())
		installedVersion := version.Semantic{}
		err := yaml.Unmarshal(itr.Value(), &installedVersion)
		if err != nil {
			return err
		}

		repoAlias, vmName := util.ParseQualifiedName(fullVMName)
		organization, repo := util.ParseAlias(repoAlias)

		recordBytes, err := u.repoRegistry.VMs().Get([]byte(vmName))
		if err != nil {
			return err
		}

		definition := repository.Definition[types.VM]{}
		if err := yaml.Unmarshal(recordBytes, &definition); err != nil {
			return err
		}
		updatedVM := definition.Definition

		if installedVersion.Compare(updatedVM.Version) < 0 {
			fmt.Printf(
				"Detected an update for %s from v%v.%v.%v to v%v.%v.%v.\n",
				fullVMName,
				installedVersion.Major,
				installedVersion.Minor,
				installedVersion.Patch,
				updatedVM.Version.Major,
				updatedVM.Version.Minor,
				updatedVM.Version.Patch,
			)
			installWorkflow := NewInstallWorkflow(InstallWorkflowConfig{
				Name:         fullVMName,
				Plugin:       vmName,
				Organization: organization,
				Repo:         repo,
				TmpPath:      u.tmpPath,
				PluginPath:   u.pluginPath,
				InstalledVMs: u.installedVMs,
				Registry:     u.repoRegistry,
				HttpClient:   u.httpClient,
			})

			fmt.Printf(
				"Rebuilding binaries for %s v%v.%v.%v.\n",
				fullVMName,
				updatedVM.Version.Major,
				updatedVM.Version.Minor,
				updatedVM.Version.Patch,
			)
			if err := u.engine.Execute(installWorkflow); err != nil {
				return err
			}

			updated = true
		}
	}

	if !updated {
		fmt.Printf("No changes detected.")
		return nil
	}

	return nil
}

func (u *UpdateWorkflow) Execute() error {
	if err := u.updateDefinitions(); err != nil {
		fmt.Printf("Unexpected error while updating definitions. %s", err)
		return err
	}

	if err := u.updateVMs(); err != nil {
		fmt.Printf("Unexpected error while updating vms. %s", err)
		return err
	}

	// checkpoint progress
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

	fmt.Printf("Finished update.\n")

	return nil
}

func (u *UpdateWorkflow) updateDefinitions() error {
	repoVMs := u.repoRegistry.VMs()
	repoSubnets := u.repoRegistry.Subnets()
	vmsPath := filepath.Join(u.repositoryPath, vmDir)

	if err := loadFromYAML[types.VM](vmKey, vmsPath, u.aliasBytes, u.latestCommit, u.globalRegistry.VMs(), repoVMs); err != nil {
		return err
	}

	subnetsPath := filepath.Join(u.repositoryPath, subnetDir)
	if err := loadFromYAML[types.Subnet](subnetKey, subnetsPath, u.aliasBytes, u.latestCommit, u.globalRegistry.Subnets(), repoSubnets); err != nil {
		return err
	}

	// Now we need to delete anything that wasn't updated in the latest commit
	if err := deleteStalePlugins[types.VM](repoVMs, u.latestCommit); err != nil {
		return err
	}
	if err := deleteStalePlugins[types.Subnet](repoSubnets, u.latestCommit); err != nil {
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
		definition := &repository.Definition[T]{
			Definition: data[key],
			Commit:     commit,
		}

		alias := data[key].Alias()
		aliasBytes := []byte(alias)

		registry := &repository.List{}
		registryBytes, err := globalDB.Get(aliasBytes)
		if err == database.ErrNotFound {
			registry = &repository.List{
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
		updatedRecordBytes, err := yaml.Marshal(definition)
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

func deleteStalePlugins[T types.Definition](db database.Database, latestCommit plumbing.Hash) error {
	itr := db.NewIterator()
	batch := db.NewBatch()

	for itr.Next() {
		definition := &repository.Definition[T]{}
		if err := yaml.Unmarshal(itr.Value(), definition); err != nil {
			return nil
		}

		if definition.Commit != latestCommit {
			fmt.Printf("Deleting a stale plugin: %s@%s as of %s.\n", definition.Definition.Alias(), definition.Commit, latestCommit)
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
