// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/types"
)

var (
	sourceInfoPrefix   = []byte("source_info")
	vmPrefix           = []byte("vm")
	subnetPrefix       = []byte("subnet")
	registryPrefix     = []byte("registry")
	installedVMsPrefix = []byte("installed_vms")

	_ Storage[any] = &Database[any]{}
)

type Storage[V any] interface {
	Has(key []byte) (bool, error)
	Put(key []byte, value V) error
	Get(key []byte) (V, error)
	Delete(key []byte) error
	Iterator() Iterator[V]
	// TODO batching
}

func NewSourceInfo(db database.Database) *Database[SourceInfo] {
	return &Database[SourceInfo]{
		db: prefixdb.New(sourceInfoPrefix, db),
	}
}

func NewVM(db database.Database) *Database[Definition[types.VM]] {
	return &Database[Definition[types.VM]]{
		db: prefixdb.New(vmPrefix, db),
	}
}

func NewSubnet(db database.Database) *Database[Definition[types.Subnet]] {
	return &Database[Definition[types.Subnet]]{
		db: prefixdb.New(subnetPrefix, db),
	}
}

func NewRegistry(db database.Database) *Database[RepoList] {
	return &Database[RepoList]{
		db: prefixdb.New(registryPrefix, db),
	}
}

func NewInstalledVMs(db database.Database) *Database[InstallInfo] {
	return &Database[InstallInfo]{
		db: prefixdb.New(installedVMsPrefix, db),
	}
}

type Database[V any] struct {
	db database.Database
}

func (c *Database[V]) Has(key []byte) (bool, error) {
	return c.db.Has(key)
}

func (c *Database[V]) Put(key []byte, value V) error {
	valueBytes, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	return c.db.Put(key, valueBytes)
}

func (c *Database[V]) Get(key []byte) (V, error) {
	value := new(V)
	valueBytes, err := c.db.Get(key)
	if err != nil {
		return *value, err
	}

	if err := yaml.Unmarshal(valueBytes, value); err != nil {
		return *value, err
	}

	return *value, nil
}

func (c *Database[V]) Delete(key []byte) error {
	return c.db.Delete(key)
}

func (c *Database[V]) Iterator() Iterator[V] {
	return Iterator[V]{
		itr: c.db.NewIterator(),
	}
}
