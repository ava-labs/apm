// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package mocks

import "github.com/ava-labs/avalanchego/database"

var _ database.Iterator = &MockDatabaseIterator{}

type MockDatabaseIterator struct {
	NextV        bool
	ErrV         error
	KeyV, ValueV []byte
}

func (m MockDatabaseIterator) Next() bool {
	return m.NextV
}

func (m MockDatabaseIterator) Error() error {
	return m.ErrV
}

func (m MockDatabaseIterator) Key() []byte {
	return m.KeyV
}

func (m MockDatabaseIterator) Value() []byte {
	return m.ValueV
}

func (m MockDatabaseIterator) Release() {
}
