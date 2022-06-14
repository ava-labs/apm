// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

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
	pluginBytes := []byte("vm")
	nameBytes := []byte("organization/repository:vm")

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
				mocks.vmStorage.EXPECT().Has(pluginBytes).Return(false, errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, errWrong, err)
			},
		},
		{
			name: "uninstalling an invalid vm",
			setup: func(mocks mocks) {
				mocks.installedVMs.EXPECT().Has(nameBytes).Return(true, nil)
				mocks.vmStorage.EXPECT().Has(pluginBytes).Return(false, nil)
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
				mocks.vmStorage.EXPECT().Has(pluginBytes).Return(true, nil)
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
				mocks.vmStorage.EXPECT().Has(pluginBytes).Return(true, nil)
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
			var installedVMs *storage.MockStorage[version.Semantic]

			vmStorage = storage.NewMockStorage[storage.Definition[types.VM]](ctrl)
			installedVMs = storage.NewMockStorage[version.Semantic](ctrl)

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
				},
			)

			test.wantErr(t, wf.Execute())
		})
	}
}
