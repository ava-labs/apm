package git

import (
	"io"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

var _ Repository = &Remote{}

type Repository interface {
	Head() (plumbing.Hash, error)
}

type Remote struct {
	repo *git.Repository
}

func NewRemote(url string, path string, reference plumbing.ReferenceName, auth *http.BasicAuth) (*Remote, error) {
	var repo *git.Repository
	if _, err := os.Stat(path); err == nil {
		// already exists, we need to check out the latest changes
		repo, err = git.PlainOpen(path)
		if err != nil {
			return nil, err
		}
		worktree, err := repo.Worktree()
		if err != nil {
			return nil, err
		}
		if err := worktree.Pull(
			//TODO use fetch + checkout instead of pull
			&git.PullOptions{
				RemoteName:    "origin",
				ReferenceName: reference,
				SingleBranch:  true,
				Auth:          auth,
				Progress:      io.Discard,
			},
		); err != nil && err != git.NoErrAlreadyUpToDate {
			return nil, err
		}
	} else if os.IsNotExist(err) {
		// otherwise, we need to clone the repository
		repo, err = git.PlainClone(path, false, &git.CloneOptions{
			URL:           url,
			ReferenceName: reference,
			SingleBranch:  true,
			Auth:          auth,
			Progress:      io.Discard,
		})
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return &Remote{repo: repo}, nil
}

func (r Remote) Head() (plumbing.Hash, error) {
	head, err := r.repo.Head()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return head.Hash(), nil
}
