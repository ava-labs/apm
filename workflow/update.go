// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/afero"

	"github.com/ava-labs/apm/git"
	"github.com/ava-labs/apm/state"
	"github.com/ava-labs/apm/util"
)

var _ Workflow = &Update{}

type UpdateConfig struct {
	Executor         Executor
	TmpPath          string
	PluginPath       string
	Installer        Installer
	RepositoriesPath string
	Auth             http.BasicAuth
	RepoFactory      state.RepositoryFactory
	Fs               afero.Fs
	StateFile        state.File
	Git              git.Factory
}

func NewUpdate(config UpdateConfig) *Update {
	return &Update{
		executor:         config.Executor,
		tmpPath:          config.TmpPath,
		pluginPath:       config.PluginPath,
		installer:        config.Installer,
		repositoriesPath: config.RepositoriesPath,
		auth:             config.Auth,
		repoFactory:      config.RepoFactory,
		fs:               config.Fs,
		stateFile:        config.StateFile,
		git:              config.Git,
	}
}

type Update struct {
	executor         Executor
	installer        Installer
	auth             http.BasicAuth
	tmpPath          string
	pluginPath       string
	repositoriesPath string
	repoFactory      state.RepositoryFactory
	fs               afero.Fs
	git              git.Factory
	stateFile        state.File
}

func (u Update) Execute() error {
	updated := 0

	fmt.Printf("Checking for updates...\n")

	for alias, sourceInfo := range u.stateFile.Sources {
		organization, repo := util.ParseAlias(alias)

		previousCommit := sourceInfo.Commit
		repositoryPath := filepath.Join(u.repositoriesPath, organization, repo)
		latestCommit, err := u.git.GetRepository(sourceInfo.URL, repositoryPath, sourceInfo.Branch, &u.auth)
		if err != nil {
			return err
		}

		if latestCommit != previousCommit {
			fmt.Printf("Updated definitions for %s@%s.\n", alias, latestCommit)
			updated++
		}

		u.stateFile.Sources[alias].Commit = latestCommit
	}

	if updated == 0 {
		fmt.Printf("All repositories are already up-to-date.\n")
	}

	return nil
}
