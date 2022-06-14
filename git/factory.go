// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package git

import (
	"io"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type Factory interface {
	GetRepository(url string, path string, reference plumbing.ReferenceName, auth *http.BasicAuth) (plumbing.Hash, error)
}

type RepositoryFactory struct{}

func (f RepositoryFactory) GetRepository(url string, path string, reference plumbing.ReferenceName, auth *http.BasicAuth) (plumbing.Hash, error) {
	var repo *git.Repository

	switch _, err := os.Stat(path); err {
	case nil:
		// already exists, so we need to check out the latest changes
		repo, err = git.PlainOpen(path)
		if err != nil {
			return plumbing.ZeroHash, err
		}
		worktree, err := repo.Worktree()
		if err != nil {
			return plumbing.ZeroHash, err
		}
		if err := worktree.Pull(
			// TODO use fetch + checkout instead of pull
			&git.PullOptions{
				RemoteName:    "origin",
				ReferenceName: reference,
				SingleBranch:  true,
				Auth:          auth,
				Progress:      io.Discard,
			},
		); err != nil && err != git.NoErrAlreadyUpToDate {
			return plumbing.ZeroHash, err
		}
	default:
		if os.IsNotExist(err) {
			// if we don't have the repo, we need to clone it
			repo, err = git.PlainClone(path, false, &git.CloneOptions{
				URL:           url,
				ReferenceName: reference,
				SingleBranch:  true,
				Auth:          auth,
				Progress:      io.Discard,
			})
			if err != nil {
				return plumbing.ZeroHash, err
			}
		} else {
			return plumbing.ZeroHash, err
		}
	}

	head, err := repo.Head()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return head.Hash(), nil
}
