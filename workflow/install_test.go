// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/ava-labs/avalanchego/utils/perms"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/apm/checksum"
	"github.com/ava-labs/apm/state"
	"github.com/ava-labs/apm/types"
)

func TestInstallExecute(t *testing.T) {
	hash := []byte("foobar")

	definition := state.Definition[types.VM]{
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
		},
		Commit: "commit",
	}
	vm := definition.Definition

	noInstallScriptDefinition := state.Definition[types.VM]{
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
		},
		Commit: "commit",
	}
	noInstallScriptVM := noInstallScriptDefinition.Definition

	installPath := filepath.Join("tmpPath", "organization", "repo")
	workingDir := filepath.Join("tmpPath", "organization", "repo", "plugin")
	tarPath := filepath.Join(installPath, "plugin.tar.gz")
	errWrong := fmt.Errorf("something went wrong")

	type mocks struct {
		stateFile   state.File
		repository  *state.MockRepository
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
			name: "download fails",
			setup: func(mocks mocks) {
				mocks.repository.EXPECT().GetVM("plugin").Return(definition, nil)
				mocks.installer.EXPECT().Download(vm.URL, tarPath).Return(errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, err, errWrong)
			},
		},
		{
			name: "wrong checksum",
			setup: func(mocks mocks) {
				mocks.repository.EXPECT().GetVM("plugin").Return(definition, nil)
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
				mocks.repository.EXPECT().GetVM("plugin").Return(definition, nil)
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
				mocks.repository.EXPECT().GetVM("plugin").Return(definition, nil)
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
				mocks.repository.EXPECT().GetVM("plugin").Return(definition, nil)
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
				mocks.repository.EXPECT().GetVM("plugin").Return(noInstallScriptDefinition, nil)
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

			stateFile, err := state.New("stateFilePath")
			require.NoError(t, err)

			installer := NewMockInstaller(ctrl)
			fs := afero.NewMemMapFs()
			checksummer := checksum.NewMockChecksummer(ctrl)
			repository := state.NewMockRepository(ctrl)

			test.setup(mocks{
				stateFile:   stateFile,
				repository:  repository,
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
					Repository:   repository,
					Fs:           fs,
					Installer:    installer,
				},
			)
			wf.checksummer = checksummer

			test.wantErr(t, wf.Execute())
		})
	}
}
