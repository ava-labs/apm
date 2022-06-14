// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"
	"testing"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/version"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/types"
)

func TestUninstallExecute(t *testing.T) {
	errWrong := fmt.Errorf("something went wrong")
	pluginBytes := []byte("vm")
	nameBytes := []byte("organization/repository:vm")

	definition := storage.Definition[types.VM]{
		Definition: types.VM{
			ID:            "id",
			Alias:         "vm",
			Homepage:      "homepage",
			Description:   "description",
			Maintainers:   []string{"joshua", "kim"},
			InstallScript: "./installScript",
			BinaryPath:    "./build/binaryPath",
			URL:           "url",
			SHA256:        "sha256",
			Version:       version.NewDefaultSemantic(1, 2, 3),
		},
		Commit: plumbing.NewHash("foobar commit"),
	}

	type mocks struct {
		vmStorage    *storage.MockStorage[storage.Definition[types.VM]]
		installedVMs *storage.MockStorage[storage.InstallInfo]
	}
	tests := []struct {
		name    string
		setup   func(mocks)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "can't read from installed vms",
			setup: func(mocks mocks) {
				mocks.installedVMs.EXPECT().Has(nameBytes).Return(false, errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, err, errWrong)
			},
		},
		{
			name: "vm already uninstalled",
			setup: func(mocks mocks) {
				mocks.installedVMs.EXPECT().Has(nameBytes).Return(false, nil)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
		{
			name: "can't read from repository vms",
			setup: func(mocks mocks) {
				mocks.installedVMs.EXPECT().Has(nameBytes).Return(true, nil)
				mocks.vmStorage.EXPECT().Get(pluginBytes).Return(storage.Definition[types.VM]{}, errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, errWrong, err)
			},
		},
		{
			name: "uninstalling an invalid vm",
			setup: func(mocks mocks) {
				mocks.installedVMs.EXPECT().Has(nameBytes).Return(true, nil)
				mocks.vmStorage.EXPECT().Get(pluginBytes).Return(storage.Definition[types.VM]{}, database.ErrNotFound)
				mocks.installedVMs.EXPECT().Delete(nameBytes).Return(nil)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
		{
			name: "removing from installation registry fails",
			setup: func(mocks mocks) {
				mocks.installedVMs.EXPECT().Has(nameBytes).Return(true, nil)
				mocks.vmStorage.EXPECT().Get(pluginBytes).Return(definition, nil)
				mocks.installedVMs.EXPECT().Delete(nameBytes).Return(errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, errWrong, err)
			},
		},
		{
			name: "success",
			setup: func(mocks mocks) {
				mocks.installedVMs.EXPECT().Has(nameBytes).Return(true, nil)
				mocks.vmStorage.EXPECT().Get(pluginBytes).Return(definition, nil)
				mocks.installedVMs.EXPECT().Delete(nameBytes).Return(nil)
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
			var installedVMs *storage.MockStorage[storage.InstallInfo]

			vmStorage = storage.NewMockStorage[storage.Definition[types.VM]](ctrl)
			installedVMs = storage.NewMockStorage[storage.InstallInfo](ctrl)

			test.setup(mocks{
				vmStorage:    vmStorage,
				installedVMs: installedVMs,
			})

			wf := NewUninstall(
				UninstallConfig{
					Name:         "organization/repository:vm",
					Plugin:       "vm",
					RepoAlias:    "organization/repository",
					VMStorage:    vmStorage,
					InstalledVMs: installedVMs,
					Fs:           afero.NewMemMapFs(),
				},
			)

			test.wantErr(t, wf.Execute())
		})
	}
}
