package storage

import "github.com/ava-labs/avalanchego/database"

type Storage[T any] interface {
	Put([]byte, T) error
	Get([]byte) (T, error)
	Delete([]byte) error
	Iterator() database.Iterator
}
