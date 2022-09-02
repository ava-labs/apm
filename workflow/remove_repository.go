// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ava-labs/apm/constant"
	"github.com/ava-labs/apm/state"
)

var _ Workflow = RemoveRepository{}

func NewRemoveRepository(config RemoveRepositoryConfig) *RemoveRepository {
	return &RemoveRepository{
		sourcesList:      config.SourcesList,
		repositoriesPath: config.RepositoriesPath,
		alias:            config.Alias,
	}
}

type RemoveRepositoryConfig struct {
	SourcesList      map[string]*state.SourceInfo
	RepositoriesPath string
	Alias            string
}

type RemoveRepository struct {
	sourcesList      map[string]*state.SourceInfo
	repositoriesPath string
	alias            string
}

func (r RemoveRepository) Execute() error {
	if r.alias == constant.CoreAlias {
		fmt.Printf("Can't remove %s (required repository).\n", constant.CoreAlias)
		return nil
	}

	_, ok := r.sourcesList[r.alias]

	repoPath := filepath.Join(r.repositoriesPath, r.alias)
	if err := os.RemoveAll(repoPath); err != nil {
		return err
	}

	if !ok {
		fmt.Printf("%s is already not a tracked repository. Skipping...\n", r.alias)
		return nil
	}

	delete(r.sourcesList, r.alias)
	fmt.Printf("Successfully removed %s\n", r.alias)
	return nil
}
