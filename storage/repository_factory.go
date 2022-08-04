// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
)

var _ RepositoryFactory = DiskRepository{}

type RepositoryFactory interface {
	GetRepository(alias []byte) Repository
}

// TODO replace with fs-based implementation
func NewRepositoryFactory(db database.Database) RepositoryFactory {
	return &DiskRepository{
		db: db,
	}
}

type DiskRepository struct {
	db database.Database
}

func (r DiskRepository) GetRepository(alias []byte) Repository {
	// all repositories
	reposDB := prefixdb.New(repositoryPrefix, r.db)
	// this specific repository
	repoDB := prefixdb.New(alias, reposDB)

	return Repository{
		VMs:     NewVM(repoDB),
		Subnets: NewSubnet(repoDB),
	}
}
