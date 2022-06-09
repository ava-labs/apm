package workflow

import (
	"testing"

	"github.com/ava-labs/avalanchego/version"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/apm/storage"
	mockable "github.com/ava-labs/apm/storage/mocks"
)

func TestUpdateRepositoryExecute(t *testing.T) {
	const (
		repoName       = "repoName"
		repositoryPath = "/path/to/repository"

		alias = "organization/repository"
		url   = "url"

		tmpPath    = "/path/to/tmp"
		pluginPath = "/path/to/pluginDir"
	)
	var (
		aliasBytes     = []byte(alias)
		previousCommit = plumbing.Hash{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		latestCommit   = plumbing.Hash{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
		sourceInfo     = storage.SourceInfo{
			Alias:  alias,
			URL:    url,
			Commit: previousCommit,
		}

		fs = afero.NewMemMapFs()
	)

	type mocks struct {
	}
	tests := []struct {
		name    string
		setup   func(mocks)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			setup: func(mocks mocks) {
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Skip()
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			executor := NewMockExecutor(ctrl)
			db := mockable.NewMockDatabase(ctrl)
			installer := NewMockInstaller(ctrl)

			var (
				registry     *storage.MockStorage[storage.RepoList]
				sourcesList  *storage.MockStorage[storage.SourceInfo]
				installedVMs *storage.MockStorage[version.Semantic]
			)

			registry = storage.NewMockStorage[storage.RepoList](ctrl)
			sourcesList = storage.NewMockStorage[storage.SourceInfo](ctrl)
			installedVMs = storage.NewMockStorage[version.Semantic](ctrl)

			wf := NewUpdateRepository(
				UpdateRepositoryConfig{
					Executor:       executor,
					RepoName:       repoName,
					RepositoryPath: repositoryPath,
					AliasBytes:     aliasBytes,
					PreviousCommit: previousCommit,
					LatestCommit:   latestCommit,
					SourceInfo:     sourceInfo,
					// Repository:     repository,
					Registry:     registry,
					SourcesList:  sourcesList,
					InstalledVMs: installedVMs,
					DB:           db,
					TmpPath:      tmpPath,
					PluginPath:   pluginPath,
					Installer:    installer,
					Fs:           fs,
				},
			)

			test.wantErr(t, wf.Execute())
		})
	}
}
