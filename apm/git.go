package apm

import (
	"io/ioutil"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func syncRepository(url string, path string, reference plumbing.ReferenceName) (*git.Repository, error) {
	var gitRepository *git.Repository
	if _, err := os.Stat(path); err == nil {
		// already exists, we need to check out the latest changes
		gitRepository, err = git.PlainOpen(path)
		if err != nil {
			return nil, err
		}
		worktree, err := gitRepository.Worktree()
		if err != nil {
			return nil, err
		}
		err = worktree.Pull(
			//TODO use fetch + checkout instead of pull
			&git.PullOptions{
				RemoteName:    "origin",
				ReferenceName: reference,
				SingleBranch:  true,
				Auth:          auth,
				Progress:      ioutil.Discard,
			},
		)
	} else if os.IsNotExist(err) {
		// otherwise, we need to clone the repository
		gitRepository, err = git.PlainClone(path, false, &git.CloneOptions{
			URL:           url,
			ReferenceName: reference,
			SingleBranch:  true,
			Auth:          auth,
			Progress:      ioutil.Discard,
		})
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return gitRepository, nil
}
