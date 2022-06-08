package storage

import (
	"github.com/ava-labs/avalanchego/database"
	"gopkg.in/yaml.v3"
)

func NewIterator[V any](itr database.Iterator) *Iterator[V] {
	return &Iterator[V]{
		itr: itr,
	}
}

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
