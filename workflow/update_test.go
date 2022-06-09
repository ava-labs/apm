package workflow

import (
	"path/filepath"
	"testing"

	"github.com/ava-labs/avalanchego/version"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/git"
	"github.com/ava-labs/apm/storage"
	mockdb "github.com/ava-labs/apm/storage/mocks"
)

func TestUpdateExecute(t *testing.T) {
	// errWrong := fmt.Errorf("something went wrong")
	//
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
		previousCommit  = plumbing.Hash{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		latestCommit    = plumbing.Hash{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
		repoInstallPath = filepath.Join(repositoriesPath, organization, repo)
		repository      = storage.Repository{}

		auth = http.BasicAuth{
			Username: "username",
			Password: "password",
		}

		sourceInfo = storage.SourceInfo{
			Alias:  alias,
			URL:    url,
			Commit: previousCommit,
		}
	)

	sourceInfoBytes, err := yaml.Marshal(sourceInfo)
	if err != nil {
		t.Fatal(err)
	}

	type mocks struct {
		ctrl         *gomock.Controller
		executor     *MockExecutor
		registry     *storage.MockStorage[storage.RepoList]
		installedVMs *storage.MockStorage[version.Semantic]
		sourcesList  *storage.MockStorage[storage.SourceInfo]
		db           *mockdb.MockDatabase
		installer    *MockInstaller
		gitFactory   *git.MockFactory
		repoFactory  *storage.MockRepositoryFactory
		auth         http.BasicAuth
	}
	tests := []struct {
		name    string
		setup   func(mocks)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success single repository",
			setup: func(mocks mocks) {

				// iterator with only one key/value pair
				mocks.sourcesList.EXPECT().Iterator().DoAndReturn(func() storage.Iterator[storage.SourceInfo] {
					itr := mockdb.NewMockIterator(mocks.ctrl)
					defer itr.EXPECT().Release()

					itr.EXPECT().Next().Return(true)
					itr.EXPECT().Key().Return([]byte(alias))

					itr.EXPECT().Value().Return(sourceInfoBytes)
					itr.EXPECT().Next().Return(false)

					return *storage.NewIterator[storage.SourceInfo](itr)
				})

				wf := NewUpdateRepository(UpdateRepositoryConfig{
					Executor:       mocks.executor,
					RepoName:       repo,
					RepositoryPath: repoInstallPath,
					AliasBytes:     []byte(alias),
					PreviousCommit: previousCommit,
					LatestCommit:   latestCommit,
					Repository:     repository,
					Registry:       mocks.registry,
					SourceInfo:     sourceInfo,
					SourcesList:    mocks.sourcesList,
					InstalledVMs:   mocks.installedVMs,
					DB:             mocks.db,
					TmpPath:        tmpPath,
					PluginPath:     pluginPath,
					Installer:      mocks.installer,
				})

				mocks.gitFactory.EXPECT().GetRepository(url, repoInstallPath, mainBranch, &mocks.auth).Return(latestCommit, nil)
				mocks.repoFactory.EXPECT().GetRepository([]byte(alias)).Return(repository)
				mocks.executor.EXPECT().Execute(wf)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			var registry *storage.MockStorage[storage.RepoList]
			var installedVMs *storage.MockStorage[version.Semantic]
			var sourcesList *storage.MockStorage[storage.SourceInfo]

			executor := NewMockExecutor(ctrl)
			db := mockdb.NewMockDatabase(ctrl)
			installer := NewMockInstaller(ctrl)
			gitFactory := git.NewMockFactory(ctrl)
			repoFactory := storage.NewMockRepositoryFactory(ctrl)

			registry = storage.NewMockStorage[storage.RepoList](ctrl)
			installedVMs = storage.NewMockStorage[version.Semantic](ctrl)
			sourcesList = storage.NewMockStorage[storage.SourceInfo](ctrl)

			test.setup(mocks{
				ctrl:         ctrl,
				executor:     executor,
				registry:     registry,
				installedVMs: installedVMs,
				sourcesList:  sourcesList,
				db:           db,
				installer:    installer,
				gitFactory:   gitFactory,
				auth:         auth,
				repoFactory:  repoFactory,
			})

			wf := NewUpdate(
				UpdateConfig{
					Executor:         executor,
					Registry:         registry,
					InstalledVMs:     installedVMs,
					SourcesList:      sourcesList,
					DB:               db,
					TmpPath:          tmpPath,
					PluginPath:       pluginPath,
					Installer:        installer,
					RepositoriesPath: repositoriesPath,
					Auth:             auth,
					GitFactory:       gitFactory,
					RepoFactory:      repoFactory,
				},
			)
			test.wantErr(t, wf.Execute())
		})
	}
}
