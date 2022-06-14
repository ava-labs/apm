// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package mocks

import "github.com/ava-labs/avalanchego/database"

var _ database.Iterator = &MockDatabaseIterator{}

type MockDatabaseIterator struct {
	Next_        bool
	Err_         error
	Key_, Value_ []byte
}

func (m MockDatabaseIterator) Next() bool {
	return m.Next_
}

func (m MockDatabaseIterator) Error() error {
	return m.Err_
}

func (m MockDatabaseIterator) Key() []byte {
	return m.Key_
}

func (m MockDatabaseIterator) Value() []byte {
	return m.Value_
}

func (m MockDatabaseIterator) Release() {
	return
}
