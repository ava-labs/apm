package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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
			mockIterator := mockDatabaseIterator{}
			mockIterator.next = test.next

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
			mockIterator := mockDatabaseIterator{}
			mockIterator.err = test.err

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
			mockIterator := mockDatabaseIterator{}
			mockIterator.key = test.key

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

	tests := []struct {
		name  string
		value Foo
	}{
		{
			name:  "value",
			value: Foo{Bar: 1},
		},
		{
			name:  "zero value",
			value: Foo{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bytes, err := yaml.Marshal(test.value)
			assert.NoError(t, err)

			mockIterator := mockDatabaseIterator{}
			mockIterator.value = bytes

			itr := Iterator[Foo]{
				itr: mockIterator,
			}

			value, err := itr.Value()
			assert.Equal(t, test.value, value)
			assert.NoError(t, err)
		})
	}
}

// No tests yet
// func TestIterator_Release(t *testing.T) {
// }
