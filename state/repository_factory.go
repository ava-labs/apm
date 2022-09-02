// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"os"
	"path/filepath"

	"github.com/ava-labs/apm/git"
)

var _ RepositoryFactory = repositoryFactory{}

type RepositoryFactory interface {
	GetRepository(alias string) (Repository, error)
}

func NewRepositoryFactory(reposPath string) RepositoryFactory {
	return &repositoryFactory{
		reposPath: reposPath,
		git:       git.RepositoryFactory{},
	}
}

type repositoryFactory struct {
	reposPath string
	git       git.Factory
}

func (r repositoryFactory) GetRepository(alias string) (Repository, error) {
	path := filepath.Join(r.reposPath, alias)
	// check that this is a valid path on disk before we return it to be used
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	return &DiskRepository{
		Git:  r.git,
		Path: filepath.Join(r.reposPath, alias),
	}, nil
}
