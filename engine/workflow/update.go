package workflow

import (
	"fmt"
	"path/filepath"

	"github.com/ava-labs/avalanchego/database"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/ava-labs/apm/git"
	"github.com/ava-labs/apm/repository"
	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/url"
	"github.com/ava-labs/apm/util"
)

var _ Workflow = &Update{}

type UpdateConfig struct {
	Executor         Executor
	GlobalRegistry   repository.Registry
	InstalledVMs     database.Database
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
		globalRegistry:   config.GlobalRegistry,
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
	globalRegistry   repository.Registry
	installedVMs     database.Database
	db               database.Database
	tmpPath          string
	pluginPath       string
	httpClient       url.Client
	sourceList       storage.Storage[storage.SourceInfo]
	repositoriesPath string
	auth             http.BasicAuth
}

func (u Update) Execute() error {
	itr := u.sourceList.Iterator()

	for itr.Next() {
		aliasBytes := itr.Key()
		organization, repo := util.ParseAlias(string(aliasBytes))

		sourceInfo, err := u.sourceList.Get(aliasBytes)
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
			RepoRegistry: repository.NewRegistry(repository.RegistryConfig{
				Alias: aliasBytes,
				DB:    u.db,
			}),
			GlobalRegistry: u.globalRegistry,
			SourceInfo:     sourceInfo,
			SourceList:     u.sourceList,
			InstalledVMs:   u.installedVMs,
			DB:             u.db,
			TmpPath:        u.tmpPath,
			PluginPath:     u.pluginPath,
			HttpClient:     u.httpClient,
		})

		if err := u.executor.Execute(workflow); err != nil {
			return err
		}
	}

	return nil
}
