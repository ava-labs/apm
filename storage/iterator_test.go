// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/storage/mocks"
)

func TestIterator_Next(t *testing.T) {
	tests := []struct {
		name string
		next bool
	}{
		{
			name: "no next",
			next: false,
		},
		{
			name: "has next",
			next: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockIterator := mocks.MockDatabaseIterator{}
			mockIterator.NextV = test.next

			itr := Iterator[any]{
				itr: mockIterator,
			}

			assert.Equal(t, test.next, itr.Next())
		})
	}
}

func TestIterator_Error(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "nil error",
			err:  nil,
		},
		{
			name: "non-nil error",
			err:  fmt.Errorf("oops"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockIterator := mocks.MockDatabaseIterator{}
			mockIterator.ErrV = test.err

			itr := Iterator[any]{
				itr: mockIterator,
			}

			assert.Equal(t, test.err, itr.Error())
		})
	}
}

func TestIterator_Key(t *testing.T) {
	tests := []struct {
		name string
		key  []byte
	}{
		{
			name: "non-nil key",
			key:  []byte("key"),
		},
		{
			name: "nil key",
			key:  nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockIterator := mocks.MockDatabaseIterator{}
			mockIterator.KeyV = test.key

			itr := Iterator[any]{
				itr: mockIterator,
			}

			assert.Equal(t, test.key, itr.Key())
		})
	}
}

func TestIterator_Value(t *testing.T) {
	type Foo struct {
		Bar int `yaml:"bar"`
	}

	// setup
	foo := Foo{Bar: 1}
	fooBytes, err := yaml.Marshal(foo)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		bytes       []byte
		expected    Foo
		expectedErr bool
	}{
		{
			name:        "expected",
			bytes:       fooBytes,
			expected:    Foo{Bar: 1},
			expectedErr: false,
		},
		{
			name:        "zero expected",
			bytes:       []byte{},
			expected:    Foo{},
			expectedErr: false,
		},
		{
			name:        "error expected",
			bytes:       []byte("asdf"),
			expected:    Foo{},
			expectedErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockIterator := mocks.MockDatabaseIterator{}
			mockIterator.ValueV = test.bytes

			itr := Iterator[Foo]{
				itr: mockIterator,
			}

			value, err := itr.Value()
			assert.Equal(t, test.expected, value)
			if test.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIterator_Release(t *testing.T) {
	itr := Iterator[any]{
		itr: mocks.MockDatabaseIterator{},
	}

	itr.Release()
}
