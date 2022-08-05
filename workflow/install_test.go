// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/ava-labs/avalanchego/utils/perms"
	"github.com/ava-labs/avalanchego/version"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/apm/checksum"
	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/types"
)

func TestInstallExecute(t *testing.T) {
	hash := []byte("foobar")

	definition := storage.Definition[types.VM]{
		Definition: types.VM{
			ID:            "id",
			Alias:         "alias",
			Homepage:      "homepage",
			Description:   "description",
			Maintainers:   []string{"joshua", "kim"},
			InstallScript: "./path/to/install/script.sh",
			BinaryPath:    "./path/to/binary",
			URL:           "www.website.com",
			SHA256:        "666f6f626172",
			Version: version.Semantic{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
		},
		Commit: plumbing.ZeroHash,
	}
	vm := definition.Definition

	noInstallScriptDefinition := storage.Definition[types.VM]{
		Definition: types.VM{
			ID:            "id",
			Alias:         "alias",
			Homepage:      "homepage",
			Description:   "description",
			Maintainers:   []string{"joshua", "kim"},
			InstallScript: "", // no install script
			BinaryPath:    "./path/to/binary",
			URL:           "www.website.com",
			SHA256:        "666f6f626172",
			Version: version.Semantic{
				Major: 5,
				Minor: 6,
				Patch: 7,
			},
		},
		Commit: plumbing.ZeroHash,
	}
	noInstallScriptVM := noInstallScriptDefinition.Definition

	installPath := filepath.Join("tmpPath", "organization", "repo")
	workingDir := filepath.Join("tmpPath", "organization", "repo", "plugin")
	tarPath := filepath.Join(installPath, "plugin.tar.gz")
	errWrong := fmt.Errorf("something went wrong")

	type mocks struct {
		stateFile   storage.StateFile
		vmStorage   *storage.MockStorage[storage.Definition[types.VM]]
		installer   *MockInstaller
		checksummer *checksum.MockChecksummer
		fs          afero.Fs
	}
	tests := []struct {
		name    string
		setup   func(mocks)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "read vm registry fails",
			setup: func(mocks mocks) {
				mocks.vmStorage.EXPECT().Get([]byte("plugin")).Return(storage.Definition[types.VM]{}, errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, err, errWrong)
			},
		},
		{
			name: "download fails",
			setup: func(mocks mocks) {
				mocks.vmStorage.EXPECT().Get([]byte("plugin")).Return(definition, nil)
				mocks.installer.EXPECT().Download(vm.URL, tarPath).Return(errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, err, errWrong)
			},
		},
		{
			name: "wrong checksum",
			setup: func(mocks mocks) {
				mocks.vmStorage.EXPECT().Get([]byte("plugin")).Return(definition, nil)
				mocks.installer.EXPECT().Download(vm.URL, tarPath).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, tarPath, nil, perms.ReadWrite)
				})
				mocks.checksummer.EXPECT().Checksum(tarPath).Return([]byte("wrong checksum"))
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "decompress fails",
			setup: func(mocks mocks) {
				mocks.vmStorage.EXPECT().Get([]byte("plugin")).Return(definition, nil)
				mocks.installer.EXPECT().Download(vm.URL, tarPath).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, tarPath, nil, perms.ReadWrite)
				})
				mocks.checksummer.EXPECT().Checksum(tarPath).Return(hash)
				mocks.installer.EXPECT().Decompress(tarPath, workingDir).Return(errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, err, errWrong)
			},
		},
		{
			name: "install fails",
			setup: func(mocks mocks) {
				mocks.vmStorage.EXPECT().Get([]byte("plugin")).Return(definition, nil)
				mocks.installer.EXPECT().Download(vm.URL, tarPath).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, tarPath, nil, perms.ReadWrite)
				})
				mocks.checksummer.EXPECT().Checksum(tarPath).Return(hash)
				mocks.installer.EXPECT().Decompress(tarPath, workingDir).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, filepath.Join(workingDir, vm.BinaryPath), nil, perms.ReadWrite)
				})
				mocks.installer.EXPECT().Install(workingDir, vm.InstallScript).Return(errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, err, errWrong)
			},
		},

		{
			name: "happy case clean install",
			setup: func(mocks mocks) {
				mocks.vmStorage.EXPECT().Get([]byte("plugin")).Return(definition, nil)
				mocks.installer.EXPECT().Download(vm.URL, tarPath).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, tarPath, nil, perms.ReadWrite)
				})
				mocks.checksummer.EXPECT().Checksum(tarPath).Return(hash)
				mocks.installer.EXPECT().Decompress(tarPath, workingDir).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, filepath.Join(workingDir, vm.BinaryPath), nil, perms.ReadWrite)
				})
				mocks.installer.EXPECT().Install(workingDir, vm.InstallScript).Return(nil)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
		{
			name: "happy case no install script",
			setup: func(mocks mocks) {
				mocks.vmStorage.EXPECT().Get([]byte("plugin")).Return(noInstallScriptDefinition, nil)
				mocks.installer.EXPECT().Download(noInstallScriptVM.URL, tarPath).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, tarPath, nil, perms.ReadWrite)
				})
				mocks.checksummer.EXPECT().Checksum(tarPath).Return(hash)
				mocks.installer.EXPECT().Decompress(tarPath, workingDir).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, filepath.Join(workingDir, noInstallScriptVM.BinaryPath), nil, perms.ReadWrite)
				})
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

			vmStorage = storage.NewMockStorage[storage.Definition[types.VM]](ctrl)
			stateFile := storage.NewEmptyStateFile("stateFilePath")
			installer := NewMockInstaller(ctrl)
			fs := afero.NewMemMapFs()
			checksummer := checksum.NewMockChecksummer(ctrl)

			test.setup(mocks{
				stateFile:   stateFile,
				vmStorage:   vmStorage,
				installer:   installer,
				fs:          fs,
				checksummer: checksummer,
			})

			wf := NewInstall(
				InstallConfig{
					Name:         "name",
					Plugin:       "plugin",
					Organization: "organization",
					Repo:         "repo",
					TmpPath:      "tmpPath",
					PluginPath:   "pluginPath",
					StateFile:    stateFile,
					VMStorage:    vmStorage,
					Fs:           fs,
					Installer:    installer,
				},
			)
			wf.checksummer = checksummer

			test.wantErr(t, wf.Execute())
		})
	}
}
