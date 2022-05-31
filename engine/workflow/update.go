package workflow

import (
	"fmt"
	"path/filepath"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/version"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/ava-labs/apm/git"
	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/url"
	"github.com/ava-labs/apm/util"
)

var _ Workflow = &Update{}

type UpdateConfig struct {
	Executor         Executor
	Registry         storage.Storage[storage.RepoList]
	InstalledVMs     storage.Storage[version.Semantic]
	DB               database.Database
	TmpPath          string
	PluginPath       string
	HttpClient       url.Client
	SourceList       storage.Storage[storage.SourceInfo]
	RepositoriesPath string
	Auth             http.BasicAuth
}

func NewUpdate(config UpdateConfig) *Update {
	return &Update{
		executor:         config.Executor,
		registry:         config.Registry,
		installedVMs:     config.InstalledVMs,
		db:               config.DB,
		tmpPath:          config.TmpPath,
		pluginPath:       config.PluginPath,
		httpClient:       config.HttpClient,
		sourceList:       config.SourceList,
		repositoriesPath: config.RepositoriesPath,
		auth:             config.Auth,
	}
}

type Update struct {
	executor         Executor
	db               database.Database
	registry         storage.Storage[storage.RepoList]
	installedVMs     storage.Storage[version.Semantic]
	sourceList       storage.Storage[storage.SourceInfo]
	httpClient       url.Client
	auth             http.BasicAuth
	tmpPath          string
	pluginPath       string
	repositoriesPath string
}

func (u Update) Execute() error {
	itr := u.sourceList.Iterator()

	for itr.Next() {
		aliasBytes := itr.Key()
		organization, repo := util.ParseAlias(string(aliasBytes))

		sourceInfo, err := itr.Value()
		if err != nil {
			return err
		}
		repositoryPath := filepath.Join(u.repositoriesPath, organization, repo)
		gitRepo, err := git.NewRemote(sourceInfo.URL, repositoryPath, "refs/heads/main", &u.auth)
		if err != nil {
			return err
		}

		previousCommit := sourceInfo.Commit
		latestCommit, err := gitRepo.Head()
		if err != nil {
			return err
		}

		if latestCommit == previousCommit {
			fmt.Printf("Already at latest for %s@%s.\n", repo, latestCommit)
			continue
		}

		workflow := NewUpdateRepository(UpdateRepositoryConfig{
			Executor:       u.executor,
			RepoName:       repo,
			RepositoryPath: repositoryPath,
			AliasBytes:     aliasBytes,
			PreviousCommit: previousCommit,
			LatestCommit:   latestCommit,
			Repository: storage.NewRepository(storage.RepositoryConfig{
				Alias: aliasBytes,
				DB:    u.db,
			}),
			Registry:     u.registry,
			SourceInfo:   sourceInfo,
			SourceList:   u.sourceList,
			InstalledVMs: u.installedVMs,
			DB:           u.db,
			TmpPath:      u.tmpPath,
			PluginPath:   u.pluginPath,
			HttpClient:   u.httpClient,
		})

		if err := u.executor.Execute(workflow); err != nil {
			return err
		}
	}

	return nil
}
