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

	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/types"
)

func TestInstallExecute(t *testing.T) {
	definition := storage.Definition[types.VM]{
		Definition: types.VM{
			ID_:           "id",
			Alias_:        "alias",
			Homepage_:     "homepage",
			Description_:  "description",
			Maintainers_:  []string{"joshua", "kim"},
			InstallScript: "./path/to/install/script.sh",
			BinaryPath:    "./path/to/binary",
			URL:           "www.website.com",
			SHA256:        "sha256hash",
			Version:       version.NewDefaultSemantic(1, 0, 0),
		},
		Commit: plumbing.ZeroHash,
	}
	vm := definition.Definition

	noInstallScriptDefinition := storage.Definition[types.VM]{
		Definition: types.VM{
			ID_:           "id",
			Alias_:        "alias",
			Homepage_:     "homepage",
			Description_:  "description",
			Maintainers_:  []string{"joshua", "kim"},
			InstallScript: "", // no install script
			BinaryPath:    "./path/to/binary",
			URL:           "www.website.com",
			SHA256:        "sha256hash",
			Version:       version.NewDefaultSemantic(1, 0, 0),
		},
		Commit: plumbing.ZeroHash,
	}
	noInstallScriptVM := noInstallScriptDefinition.Definition

	installPath := filepath.Join("tmpPath", "organization", "repo")
	workingDir := filepath.Join("tmpPath", "organization", "repo", "plugin")
	tarPath := filepath.Join(installPath, "plugin.tar.gz")
	errWrong := fmt.Errorf("something went wrong")

	type mocks struct {
		installedVMs *storage.MockStorage[version.Semantic]
		vmStorage    *storage.MockStorage[storage.Definition[types.VM]]
		installer    *MockInstaller
		fs           afero.Fs
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
			name: "decompress fails",
			setup: func(mocks mocks) {
				mocks.vmStorage.EXPECT().Get([]byte("plugin")).Return(definition, nil)
				mocks.installer.EXPECT().Download(vm.URL, tarPath).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, tarPath, nil, perms.ReadWrite)
				})
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
			name: "installation registry fails",
			setup: func(mocks mocks) {
				mocks.vmStorage.EXPECT().Get([]byte("plugin")).Return(definition, nil)
				mocks.installer.EXPECT().Download(vm.URL, tarPath).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, tarPath, nil, perms.ReadWrite)
				})
				mocks.installer.EXPECT().Decompress(tarPath, workingDir).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, filepath.Join(workingDir, vm.BinaryPath), nil, perms.ReadWrite)
				})
				mocks.installer.EXPECT().Install(workingDir, vm.InstallScript).Return(nil)
				mocks.installedVMs.EXPECT().Put([]byte("name"), vm.Version).Return(errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, errWrong, err)
			},
		},
		{
			name: "happy case clean install",
			setup: func(mocks mocks) {
				mocks.vmStorage.EXPECT().Get([]byte("plugin")).Return(definition, nil)
				mocks.installer.EXPECT().Download(vm.URL, tarPath).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, tarPath, nil, perms.ReadWrite)
				})
				mocks.installer.EXPECT().Decompress(tarPath, workingDir).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, filepath.Join(workingDir, vm.BinaryPath), nil, perms.ReadWrite)
				})
				mocks.installer.EXPECT().Install(workingDir, vm.InstallScript).Return(nil)
				mocks.installedVMs.EXPECT().Put([]byte("name"), vm.Version).Return(nil)
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
				mocks.installer.EXPECT().Decompress(tarPath, workingDir).Do(func(string, string) error {
					return afero.WriteFile(mocks.fs, filepath.Join(workingDir, noInstallScriptVM.BinaryPath), nil, perms.ReadWrite)
				})
				mocks.installedVMs.EXPECT().Put([]byte("name"), noInstallScriptVM.Version).Return(nil)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			var (
				installedVMs *storage.MockStorage[version.Semantic]
				vmStorage    *storage.MockStorage[storage.Definition[types.VM]]
			)

			installedVMs = storage.NewMockStorage[version.Semantic](ctrl)
			vmStorage = storage.NewMockStorage[storage.Definition[types.VM]](ctrl)
			installer := NewMockInstaller(ctrl)
			fs := afero.NewMemMapFs()

			test.setup(mocks{
				installedVMs: installedVMs,
				vmStorage:    vmStorage,
				installer:    installer,
				fs:           fs,
			})

			wf := NewInstall(
				InstallConfig{
					Name:         "name",
					Plugin:       "plugin",
					Organization: "organization",
					Repo:         "repo",
					TmpPath:      "tmpPath",
					PluginPath:   "pluginPath",
					InstalledVMs: installedVMs,
					VMStorage:    vmStorage,
					Fs:           fs,
					Installer:    installer,
				},
			)

			test.wantErr(t, wf.Execute())
		})
	}
}
