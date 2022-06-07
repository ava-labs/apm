package workflow

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/ava-labs/apm/storage"
)

var _ Workflow = AddRepository{}

func NewAddRepository(config AddRepositoryConfig) *AddRepository {
	return &AddRepository{
		sourceList: config.SourceList,
		alias:      config.Alias,
		url:        config.Url,
	}
}

type AddRepositoryConfig struct {
	SourceList storage.Storage[storage.SourceInfo]
	Alias, Url string
}

type AddRepository struct {
	sourceList storage.Storage[storage.SourceInfo]
	alias, url string
}

func (a AddRepository) Execute() error {
	aliasBytes := []byte(a.alias)

	if ok, err := a.sourceList.Has(aliasBytes); err != nil {
		return err
	} else if ok {
		return fmt.Errorf("%s is already registered as a repository.\n", a.alias)
	}

	unsynced := storage.SourceInfo{
		Alias:  a.alias,
		URL:    a.url,
		Commit: plumbing.ZeroHash, // hasn't been synced yet
	}
	return a.sourceList.Put(aliasBytes, unsynced)
}
