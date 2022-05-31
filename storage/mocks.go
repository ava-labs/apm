package storage

import "github.com/ava-labs/avalanchego/database"

var _ database.Iterator = &mockDatabaseIterator{}

type mockDatabaseIterator struct {
	next       bool
	err        error
	key, value []byte
}

func (m mockDatabaseIterator) Next() bool {
	return m.next
}

func (m mockDatabaseIterator) Error() error {
	return m.err
}

func (m mockDatabaseIterator) Key() []byte {
	return m.key
}

func (m mockDatabaseIterator) Value() []byte {
	return m.value
}

func (m mockDatabaseIterator) Release() {
	return
}
