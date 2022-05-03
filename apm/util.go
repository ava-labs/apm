package apm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ava-labs/avalanchego/database"
	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v2"

	"github.com/ava-labs/apm/repository"
	"github.com/ava-labs/apm/types"
)

const (
	qualifiedNameDelimiter = ":"
	aliasDelimiter         = "/"
)

func parseQualifiedName(name string) (source string, plugin string) {
	parsed := strings.Split(name, qualifiedNameDelimiter)

	return parsed[0], parsed[1]
}

func parseAlias(alias string) (organization string, repository string) {
	parsed := strings.Split(alias, aliasDelimiter)

	return parsed[0], parsed[1]
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
