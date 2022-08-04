// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

var (
	vmPrefix     = []byte("vm")
	subnetPrefix = []byte("subnet")

	_ Storage[any] = &Database[any]{}
)

type Storage[V any] interface {
	Has(key []byte) (bool, error)
	Put(key []byte, value V) error
	Get(key []byte) (V, error)
	Delete(key []byte) error
	Iterator() Iterator[V]
}
