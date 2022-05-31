package storage

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/version"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/types"
)

var (
	sourceInfoPrefix   = []byte("source_info")
	vmPrefix           = []byte("vm")
	subnetPrefix       = []byte("subnet")
	registryPrefix     = []byte("registry")
	installedVMsPrefix = []byte("installed_vms")

	_ Storage[any] = &storage[any]{}
)

type Iterator[V any] struct {
	itr database.Iterator
}

func (i *Iterator[V]) Next() bool {
	return i.itr.Next()
}

func (i *Iterator[V]) Error() error {
	return i.itr.Error()
}

func (i *Iterator[V]) Key() []byte {
	return i.itr.Key()
}

func (i *Iterator[V]) Value() (V, error) {
	result := new(V)

	if err := yaml.Unmarshal(i.itr.Value(), result); err != nil {
		return *result, err
	}

	return *result, nil
}

func (i *Iterator[V]) Release() {
	i.itr.Release()
}

type Storage[V any] interface {
	Has([]byte) (bool, error)
	Put([]byte, V) error
	Get([]byte) (V, error)
	Delete([]byte) error
	Iterator() Iterator[V]
	// TODO batching
}

func NewSourceInfo(db database.Database) Storage[SourceInfo] {
	return &storage[SourceInfo]{
		db: prefixdb.New(sourceInfoPrefix, db),
	}
}

func NewVM(db database.Database) Storage[Definition[types.VM]] {
	return &storage[Definition[types.VM]]{
		db: prefixdb.New(vmPrefix, db),
	}
}

func NewSubnet(db database.Database) Storage[Definition[types.Subnet]] {
	return &storage[Definition[types.Subnet]]{
		db: prefixdb.New(subnetPrefix, db),
	}
}

func NewRegistry(db database.Database) Storage[RepoList] {
	return &storage[RepoList]{
		db: prefixdb.New(registryPrefix, db),
	}
}

func NewInstalledVMs(db database.Database) Storage[version.Semantic] {
	return &storage[version.Semantic]{
		db: prefixdb.New(installedVMsPrefix, db),
	}
}

type storage[V any] struct {
	db database.Database
}

func (c *storage[V]) Has(key []byte) (bool, error) {
	return c.db.Has(key)
}

func (c *storage[V]) Put(key []byte, value V) error {
	valueBytes, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	return c.db.Put(key, valueBytes)
}

func (c *storage[V]) Get(key []byte) (V, error) {
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

func (c *storage[V]) Delete(key []byte) error {
	return c.db.Delete(key)
}

func (c *storage[V]) Iterator() Iterator[V] {
	return Iterator[V]{
		itr: c.db.NewIterator(),
	}
}
