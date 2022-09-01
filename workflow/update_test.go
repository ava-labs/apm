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
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/apm/git"
	"github.com/ava-labs/apm/state"
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

		previousCommit  = "old"
		latestCommit    = "new"
		repoInstallPath = filepath.Join(repositoriesPath, organization, repo)

		auth = http.BasicAuth{
			Username: "username",
			Password: "password",
		}

		branch = plumbing.NewBranchReferenceName("branch")

		outdated = &state.SourceInfo{
			URL:    url,
			Branch: branch,
			Commit: previousCommit,
		}
		updated = &state.SourceInfo{
			URL:    url,
			Branch: branch,
			Commit: latestCommit,
		}

		fs = afero.NewMemMapFs()
	)

	type mocks struct {
		ctrl        *gomock.Controller
		executor    *MockExecutor
		stateFile   state.File
		installer   *MockInstaller
		git         *git.MockFactory
		repoFactory *state.MockRepositoryFactory
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
				mocks.stateFile.Sources[alias] = updated
				mocks.git.EXPECT().GetRepository(url, repoInstallPath, branch, &mocks.auth).Return("", errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, errWrong, err)
			},
		},

		{
			name: "success single repository no upgrade needed",
			setup: func(mocks mocks) {
				mocks.stateFile.Sources[alias] = updated
				mocks.git.EXPECT().GetRepository(url, repoInstallPath, branch, &mocks.auth).Return(previousCommit, nil)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "success single repository updates",
			setup: func(mocks mocks) {
				mocks.stateFile.Sources[alias] = outdated
				mocks.git.EXPECT().GetRepository(url, repoInstallPath, branch, &mocks.auth).Return(latestCommit, nil)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, nil, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			executor := NewMockExecutor(ctrl)
			installer := NewMockInstaller(ctrl)
			git := git.NewMockFactory(ctrl)
			repoFactory := state.NewMockRepositoryFactory(ctrl)

			stateFile, err := state.New("stateFilePath")
			require.NoError(t, err)

			test.setup(mocks{
				ctrl:        ctrl,
				executor:    executor,
				stateFile:   stateFile,
				installer:   installer,
				git:         git,
				auth:        auth,
				repoFactory: repoFactory,
			})

			wf := NewUpdate(
				UpdateConfig{
					Executor:         executor,
					StateFile:        stateFile,
					TmpPath:          tmpPath,
					PluginPath:       pluginPath,
					Installer:        installer,
					RepositoriesPath: repositoriesPath,
					Auth:             auth,
					Git:              git,
					RepoFactory:      repoFactory,
					Fs:               fs,
				},
			)
			test.wantErr(t, wf.Execute())
		})
	}
}
