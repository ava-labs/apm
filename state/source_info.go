// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/ava-labs/apm/types"
)

// SourceInfo represents a repository, its source, and the last synced commit.
type SourceInfo struct {
	URL    string                 `yaml:"url"`
	Commit string                 `yaml:"commit"`
	Branch plumbing.ReferenceName `yaml:"branch"`
}

type InstallInfo struct {
	ID     string `yaml:"id"`
	Commit string `yaml:"commit"`
}

// Definition stores a plugin definition alongside the plugin-repository's commit
// it was downloaded from.
type Definition[T types.Definition] struct {
	Definition T      `yaml:"definition"`
	Commit     string `yaml:"commit"`
}
