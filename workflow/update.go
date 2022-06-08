package workflow

import (
	"fmt"
	"path/filepath"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/version"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/ava-labs/apm/git"
	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/util"
)

var (
	_ Workflow = &Update{}

	mainBranch = plumbing.NewBranchReferenceName("main")
)

type UpdateConfig struct {
	Executor         Executor
	Registry         storage.Storage[storage.RepoList]
	InstalledVMs     storage.Storage[version.Semantic]
	SourcesList      storage.Storage[storage.SourceInfo]
	DB               database.Database
	TmpPath          string
	PluginPath       string
	Installer        Installer
	RepositoriesPath string
	Auth             http.BasicAuth
	GitFactory       git.Factory
	RepoFactory      storage.RepositoryFactory
}

func NewUpdate(config UpdateConfig) *Update {
	return &Update{
		executor:         config.Executor,
		registry:         config.Registry,
		installedVMs:     config.InstalledVMs,
		db:               config.DB,
		tmpPath:          config.TmpPath,
		pluginPath:       config.PluginPath,
		installer:        config.Installer,
		sourcesList:      config.SourcesList,
		repositoriesPath: config.RepositoriesPath,
		auth:             config.Auth,
		gitFactory:       config.GitFactory,
		repoFactory:      config.RepoFactory,
	}
}

type Update struct {
	executor         Executor
	db               database.Database
	registry         storage.Storage[storage.RepoList]
	installedVMs     storage.Storage[version.Semantic]
	sourcesList      storage.Storage[storage.SourceInfo]
	installer        Installer
	auth             http.BasicAuth
	tmpPath          string
	pluginPath       string
	repositoriesPath string
	gitFactory       git.Factory
	repoFactory      storage.RepositoryFactory
}

func (u Update) Execute() error {
	itr := u.sourcesList.Iterator()
	defer itr.Release()

	for itr.Next() {
		aliasBytes := itr.Key()
		organization, repo := util.ParseAlias(string(aliasBytes))

		sourceInfo, err := itr.Value()
		if err != nil {
			return err
		}
		repositoryPath := filepath.Join(u.repositoriesPath, organization, repo)
		latestCommit, err := u.gitFactory.GetRepository(sourceInfo.URL, repositoryPath, mainBranch, &u.auth)
		if err != nil {
			return err
		}

		previousCommit := sourceInfo.Commit
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
			Repository:     u.repoFactory.GetRepository(aliasBytes),
			Registry:       u.registry,
			SourceInfo:     sourceInfo,
			SourcesList:    u.sourcesList,
			InstalledVMs:   u.installedVMs,
			DB:             u.db,
			TmpPath:        u.tmpPath,
			PluginPath:     u.pluginPath,
			Installer:      u.installer,
		})

		if err := u.executor.Execute(workflow); err != nil {
			return err
		}
	}

	return nil
}