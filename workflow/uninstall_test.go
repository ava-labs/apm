package workflow

import (
	"fmt"
	"testing"

	"github.com/ava-labs/avalanchego/version"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/types"
)

func TestUninstallExecute(t *testing.T) {
	errWrong := fmt.Errorf("something went wrong")

	type mocks struct {
		vmStorage    *storage.MockStorage[storage.Definition[types.VM]]
		installedVMs *storage.MockStorage[version.Semantic]
	}
	tests := []struct {
		name    string
		setup   func(mocks)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "can't read from installed vms",
			setup: func(mocks mocks) {
				mocks.installedVMs.EXPECT().Has([]byte("name")).Return(false, errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, err, errWrong)
			},
		},
		{
			name: "vm already uninstalled",
			setup: func(mocks mocks) {
				mocks.installedVMs.EXPECT().Has([]byte("name")).Return(false, nil)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
		{
			name: "can't read from repository vms",
			setup: func(mocks mocks) {
				mocks.installedVMs.EXPECT().Has([]byte("name")).Return(true, nil)
				mocks.vmStorage.EXPECT().Has([]byte("plugin")).Return(false, errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, errWrong, err)
			},
		},
		{
			name: "uninstalling an invalid vm",
			setup: func(mocks mocks) {
				mocks.installedVMs.EXPECT().Has([]byte("name")).Return(true, nil)
				mocks.vmStorage.EXPECT().Has([]byte("plugin")).Return(false, nil)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "removing from installation registry fails",
			setup: func(mocks mocks) {
				mocks.installedVMs.EXPECT().Has([]byte("name")).Return(true, nil)
				mocks.vmStorage.EXPECT().Has([]byte("plugin")).Return(true, nil)
				mocks.installedVMs.EXPECT().Delete([]byte("name")).Return(errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, errWrong, err)
			},
		},
		{
			name: "success",
			setup: func(mocks mocks) {
				mocks.installedVMs.EXPECT().Has([]byte("name")).Return(true, nil)
				mocks.vmStorage.EXPECT().Has([]byte("plugin")).Return(true, nil)
				mocks.installedVMs.EXPECT().Delete([]byte("name")).Return(nil)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			var vmStorage *storage.MockStorage[storage.Definition[types.VM]]
			var installedVMs *storage.MockStorage[version.Semantic]

			vmStorage = storage.NewMockStorage[storage.Definition[types.VM]](ctrl)
			installedVMs = storage.NewMockStorage[version.Semantic](ctrl)

			test.setup(mocks{
				vmStorage:    vmStorage,
				installedVMs: installedVMs,
			})

			wf := NewUninstall(
				UninstallConfig{
					Name:         "name",
					Plugin:       "plugin",
					RepoAlias:    "repoAlias",
					VMStorage:    vmStorage,
					InstalledVMs: installedVMs,
				},
			)

			test.wantErr(t, wf.Execute())
		})
	}
}
