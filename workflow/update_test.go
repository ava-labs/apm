// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/apm/git"
	"github.com/ava-labs/apm/storage"
	mockdb "github.com/ava-labs/apm/storage/mocks"
)

func TestUpdateExecute(t *testing.T) {
	const (
		organization     = "organization"
		repo             = "repository"
		alias            = "organization/repository"
		url              = "url"
		tmpPath          = "tmpPath"
		pluginPath       = "pluginPath"
		repositoriesPath = "repositoriesPath"
	)

	var (
		errWrong = fmt.Errorf("something went wrong")

		previousCommit  = plumbing.Hash{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		latestCommit    = plumbing.Hash{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
		repoInstallPath = filepath.Join(repositoriesPath, organization, repo)
		repository      = storage.Repository{}

		auth = http.BasicAuth{
			Username: "username",
			Password: "password",
		}

		branch = plumbing.NewBranchReferenceName("branch")

		sourceInfo = storage.SourceInfo{
			Alias:  alias,
			URL:    url,
			Branch: branch,
			Commit: previousCommit,
		}

		fs = afero.NewMemMapFs()
	)

	type mocks struct {
		ctrl        *gomock.Controller
		executor    *MockExecutor
		stateFile   storage.StateFile
		db          *mockdb.MockDatabase
		installer   *MockInstaller
		gitFactory  *git.MockFactory
		repoFactory *storage.MockRepositoryFactory
		auth        http.BasicAuth
	}
	tests := []struct {
		name    string
		setup   func(mocks)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "cant get latest git head",
			setup: func(mocks mocks) {
				// iterator with only one key/value pair
				mocks.stateFile.Sources[alias] = sourceInfo
				mocks.gitFactory.EXPECT().GetRepository(url, repoInstallPath, branch, &mocks.auth).Return(plumbing.ZeroHash, errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, errWrong, err)
			},
		},
		{
			name: "workflow fails",
			setup: func(mocks mocks) {
				// iterator with only one key/value pair
				mocks.stateFile.Sources[alias] = sourceInfo
				wf := NewUpdateRepository(UpdateRepositoryConfig{
					RepoName:       repo,
					RepositoryPath: repoInstallPath,
					AliasBytes:     []byte(alias),
					PreviousCommit: previousCommit,
					LatestCommit:   latestCommit,
					Repository:     repository,
					SourceInfo:     sourceInfo,
					StateFile:      mocks.stateFile,
					Fs:             fs,
				})

				mocks.gitFactory.EXPECT().GetRepository(url, repoInstallPath, branch, &mocks.auth).Return(latestCommit, nil)
				mocks.repoFactory.EXPECT().GetRepository([]byte(alias)).Return(repository)
				mocks.executor.EXPECT().Execute(wf).Return(errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, errWrong, err)
			},
		},
		{
			name: "success single repository no upgrade needed",
			setup: func(mocks mocks) {
				// iterator with only one key/value pair
				mocks.stateFile.Sources[alias] = sourceInfo
				mocks.gitFactory.EXPECT().GetRepository(url, repoInstallPath, branch, &mocks.auth).Return(previousCommit, nil)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "success single repository updates",
			setup: func(mocks mocks) {
				// iterator with only one key/value pair
				mocks.stateFile.Sources[alias] = sourceInfo
				wf := NewUpdateRepository(UpdateRepositoryConfig{
					RepoName:       repo,
					RepositoryPath: repoInstallPath,
					AliasBytes:     []byte(alias),
					PreviousCommit: previousCommit,
					LatestCommit:   latestCommit,
					Repository:     repository,
					StateFile:      mocks.stateFile,
					SourceInfo:     sourceInfo,
					Fs:             fs,
				})

				mocks.gitFactory.EXPECT().GetRepository(url, repoInstallPath, branch, &mocks.auth).Return(latestCommit, nil)
				mocks.repoFactory.EXPECT().GetRepository([]byte(alias)).Return(repository)
				mocks.executor.EXPECT().Execute(wf).Return(nil)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			executor := NewMockExecutor(ctrl)
			db := mockdb.NewMockDatabase(ctrl)
			installer := NewMockInstaller(ctrl)
			gitFactory := git.NewMockFactory(ctrl)
			repoFactory := storage.NewMockRepositoryFactory(ctrl)

			stateFile := storage.NewEmptyStateFile("stateFilePath")

			test.setup(mocks{
				ctrl:        ctrl,
				executor:    executor,
				stateFile:   stateFile,
				db:          db,
				installer:   installer,
				gitFactory:  gitFactory,
				auth:        auth,
				repoFactory: repoFactory,
			})

			wf := NewUpdate(
				UpdateConfig{
					Executor:         executor,
					StateFile:        stateFile,
					DB:               db,
					TmpPath:          tmpPath,
					PluginPath:       pluginPath,
					Installer:        installer,
					RepositoriesPath: repositoriesPath,
					Auth:             auth,
					GitFactory:       gitFactory,
					RepoFactory:      repoFactory,
					Fs:               fs,
				},
			)
			test.wantErr(t, wf.Execute())
		})
	}
}
