package storage

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/types"
)

type Database[V any] struct {
	db database.Database
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
