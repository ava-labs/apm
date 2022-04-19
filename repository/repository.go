package repository

import (
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/ava-labs/apm/types"
)

type Metadata struct {
	Alias  string        `yaml:"alias"`
	URL    string        `yaml:"url"`
	Commit plumbing.Hash `yaml:"commit"`
}

type Registry struct {
	Repositories []string `yaml:"repositories"`
}

type Record[T types.Plugin] struct {
	Plugin T             `yaml:"plugin"`
	Commit plumbing.Hash `yaml:"commit"`
}
