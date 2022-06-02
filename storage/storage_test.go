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
			name: "error",
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

			db.EXPECT().Put(key, bytes).Return(test.err)

			storage := &Database[string]{
				db: db,
			}

			err = storage.Put(key, value)
			assert.Equal(t, test.err, err)
		})
	}
}

func TestDatabase_Get(t *testing.T) {
	bytes, err := yaml.Marshal("foo")
	assert.NoError(t, err)

	tests := []struct {
		name  string
		bytes []byte
		value string
		err   error
	}{
		{
			name:  "no error",
			bytes: bytes,
			value: "foo",
			err:   nil,
		},

		{
			name:  "error",
			bytes: nil,
			value: "",
			err:   fmt.Errorf("foo"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := mocks.NewMockDatabase(ctrl)

			key := []byte("key")
			db.EXPECT().Get(key).Return(bytes, test.err)

			storage := &Database[string]{
				db: db,
			}

			value, err := storage.Get(key)
			assert.Equal(t, test.value, value)
			assert.Equal(t, test.err, err)
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

func TestDatabase_Iterator(t *testing.T) {
	t.Skip("Not tested yet.")
}
