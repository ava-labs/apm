// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/storage/mocks"
)

func TestDatabase_Has(t *testing.T) {
	tests := []struct {
		name string
		has  bool
		err  error
	}{
		{
			name: "no error",
			has:  true,
			err:  nil,
		},

		{
			name: "error",
			has:  false,
			err:  fmt.Errorf("foo"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := mocks.NewMockDatabase(ctrl)

			key := []byte("key")

			db.EXPECT().Has(key).Return(test.has, test.err)

			storage := &Database[string]{
				db: db,
			}

			ok, err := storage.Has(key)
			assert.Equal(t, test.has, ok)
			assert.Equal(t, test.err, err)
		})
	}
}

func TestDatabase_Put(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "no error",
			err:  nil,
		},
		{
			name: "error on marshal",
			err:  fmt.Errorf("foo"),
		},
		{
			name: "error on put",
			err:  fmt.Errorf("foo"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := mocks.NewMockDatabase(ctrl)

			key := []byte("key")
			value := "value"
			bytes, err := yaml.Marshal(&value)
			assert.Nil(t, err)

			db.EXPECT().Put(key, bytes).Return(test.err)

			storage := &Database[string]{
				db: db,
			}

			err = storage.Put(key, value)
			assert.Equal(t, test.err, err)
		})
	}
}

// TODO verify type of error?
func TestDatabase_Get(t *testing.T) {
	type value struct {
		Foo string  `yaml:"foo"`
		Bar float32 `yaml:"bar"`
	}

	bytes, err := yaml.Marshal(value{"foo", 3.14})
	assert.NoError(t, err)

	tests := []struct {
		name      string
		bytes     []byte
		value     value
		dbErr     error
		shouldErr bool
	}{
		{
			name:      "no error",
			bytes:     bytes,
			value:     value{"foo", 3.14},
			dbErr:     nil,
			shouldErr: false,
		},

		{
			name:      "error from db",
			bytes:     nil,
			value:     value{},
			dbErr:     fmt.Errorf("foo"),
			shouldErr: true,
		},

		{
			name:      "error from deserialization",
			bytes:     []byte("wrong"),
			value:     value{},
			dbErr:     nil,
			shouldErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := mocks.NewMockDatabase(ctrl)

			key := []byte("key")
			db.EXPECT().Get(key).Return(test.bytes, test.dbErr)

			storage := &Database[value]{
				db: db,
			}

			value, err := storage.Get(key)
			assert.Equal(t, test.value, value)
			if test.shouldErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDatabase_Delete(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "no error",
			err:  nil,
		},

		{
			name: "error",
			err:  fmt.Errorf("foo"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := mocks.NewMockDatabase(ctrl)

			key := []byte("key")

			db.EXPECT().Delete(key).Return(test.err)

			storage := &Database[string]{
				db: db,
			}

			err := storage.Delete(key)
			assert.Equal(t, test.err, err)
		})
	}
}
