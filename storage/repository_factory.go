// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
)

var _ RepositoryFactory = repositoryFactory{}

type RepositoryFactory interface {
	GetRepository(alias []byte) Repository
}

func NewRepositoryFactory(db database.Database) RepositoryFactory {
	return &repositoryFactory{
		db: db,
	}
}

type repositoryFactory struct {
	db database.Database
}

func (r repositoryFactory) GetRepository(alias []byte) Repository {
	// all repositories
	reposDB := prefixdb.New(repositoryPrefix, r.db)
	// this specific repository
	repoDB := prefixdb.New(alias, reposDB)

	return Repository{
		VMs:     NewVM(repoDB),
		Subnets: NewSubnet(repoDB),
	}
}
