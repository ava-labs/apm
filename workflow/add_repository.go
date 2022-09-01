// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/ava-labs/apm/state"
)

var _ Workflow = AddRepository{}

func NewAddRepository(config AddRepositoryConfig) *AddRepository {
	return &AddRepository{
		sourcesList: config.SourcesList,
		alias:       config.Alias,
		url:         config.URL,
		branch:      config.Branch,
	}
}

type AddRepositoryConfig struct {
	SourcesList map[string]*state.SourceInfo
	Alias, URL  string
	Branch      plumbing.ReferenceName
}

type AddRepository struct {
	sourcesList map[string]*state.SourceInfo
	alias, url  string
	branch      plumbing.ReferenceName
}

func (a AddRepository) Execute() error {
	if _, ok := a.sourcesList[a.alias]; ok {
		return fmt.Errorf("%s is already registered as a repository", a.alias)
	}

	unsynced := &state.SourceInfo{
		URL:    a.url,
		Branch: a.branch,
		Commit: plumbing.ZeroHash.String(), // hasn't been synced yet
	}

	a.sourcesList[a.alias] = unsynced
	return nil
}
