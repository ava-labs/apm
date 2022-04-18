package repository

import "github.com/go-git/go-git/v5/plumbing"

type Metadata struct {
	Alias  string        `json:"alias"`
	URL    string        `json:"url"`
	Commit plumbing.Hash `json:"commit"`
}
