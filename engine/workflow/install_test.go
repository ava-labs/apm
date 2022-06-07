package workflow

import (
	"path/filepath"
	"testing"

	"github.com/ava-labs/avalanchego/version"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/types"
	"github.com/ava-labs/apm/url"
)

func TestExecute(t *testing.T) {
	definition := storage.Definition[types.VM]{
		Definition: types.VM{
			ID_:           "id",
			Alias_:        "alias",
			Homepage_:     "homepage",
			Description_:  "description",
			Maintainers_:  []string{"joshua", "kim"},
			InstallScript: "./path/to/install/script.sh",
			BinaryPath:    "./path/to/binary",
			URL:           "www.subnetsgonewild.com",
			SHA256:        "sha256hash",
			Version:       version.NewDefaultSemantic(1, 0, 0),
		},
		Commit: plumbing.ZeroHash,
	}

	vm := definition.Definition

	// errWrong := fmt.Errorf("something went wrong")

	tests := []struct {
		name  string
		setup func(
			*InstallWorkflow,
			*storage.MockStorage[version.Semantic],
			*storage.MockStorage[storage.Definition[types.VM]],
			*url.MockClient,
			afero.Fs,
		)
		err error
	}{
		{
			name: "no prior install and no errors",
			setup: func(
				wf *InstallWorkflow,
				installedVMs *storage.MockStorage[version.Semantic],
				vmStorage *storage.MockStorage[storage.Definition[types.VM]],
				urlClient *url.MockClient,
				fs afero.Fs,
			) {
				installPath := filepath.Join(wf.tmpPath, wf.organization, wf.repo)

				vmStorage.EXPECT().Get([]byte("plugin")).Return(definition, nil)
				urlClient.EXPECT().Download(filepath.Join(installPath, "plugin.tar.gz"), vm.URL).Return(nil)
			},
			err: nil,
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
			urlClient := url.NewMockClient(ctrl)
			fs := afero.NewMemMapFs()

			wf := NewInstallWorkflow(
				InstallWorkflowConfig{
					Name:         "name",
					Plugin:       "plugin",
					Organization: "organization",
					Repo:         "repo",
					TmpPath:      "tmpPath",
					PluginPath:   "pluginPath",
					InstalledVMs: installedVMs,
					VMStorage:    vmStorage,
					UrlClient:    urlClient,
					Fs:           fs,
					Installer:    NewMockInstaller(ctrl),
				},
			)

			test.setup(wf, installedVMs, vmStorage, urlClient, fs)

			err := wf.Execute()
			assert.Equal(t, test.err, err)
		})
	}
}
