package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/leveldb"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/utils/filesystem"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/perms"
	"github.com/ava-labs/avalanchego/utils/subprocess"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/go-plugin"

	"github.com/ava-labs/avalanche-plugin/constant"
	"github.com/ava-labs/avalanche-plugin/grpc"
	avaxPlugin "github.com/ava-labs/avalanche-plugin/plugin"
	"github.com/ava-labs/avalanche-plugins-core/core"
)

var (
	dbDir           = "db"
	repositoriesDir = "repositories"
	buildDir        = "build"

	repoPrefix   = []byte("repo")
	syncedPrefix = []byte("synced")

	codecVersion uint16 = 1

	auth = &http.BasicAuth{
		Username: "personal access token",
		//TODO accept token through cli
		Password: "<YOUR PERSONAL ACCESS TOKEN HERE>",
	}
)

type Service struct {
	repositoriesPath string
	buildPath        string

	codecManager codec.Manager
	db           database.Database
	//TODO merge these databases together
	repositoriesDB database.Database
	syncedDB       database.Database
	fsReader       filesystem.Reader
}

func (s *Service) Install(alias string) error {
	return nil
}

func (s *Service) Uninstall(alias string) error {
	return nil
}

func (s *Service) Upgrade(alias string) error {
	return nil
}

func (s *Service) Search(alias string) error {
	return nil
}

func (s *Service) Info(alias string) error {
	return nil
}

func (s *Service) Update() error {
	itr := s.repositoriesDB.NewIterator()
	for itr.Next() {
		// Need to split the alias to support Windows
		aliasBytes := itr.Key()
		alias := string(aliasBytes)
		aliasSplit := strings.Split(alias, "/")
		organizationName := aliasSplit[0]
		repositoryName := aliasSplit[1]

		repositoryURL := string(itr.Value())
		repositoryPath := filepath.Join(s.repositoriesPath, organizationName, repositoryName)

		var gitRepository *git.Repository

		if _, err := os.Stat(repositoryPath); err == nil {
			// already exists, we need to check out the latest changes
			gitRepository, err = git.PlainOpen(repositoryPath)
			if err != nil {
				return err
			}
			worktree, err := gitRepository.Worktree()
			if err != nil {
				return err
			}
			err = worktree.Pull(
				//TODO use fetch + checkout instead of pull
				&git.PullOptions{
					RemoteName:    "origin",
					Auth:          auth,
					ReferenceName: "refs/heads/testing",
					SingleBranch:  true,
				},
			)
		} else if os.IsNotExist(err) {
			// otherwise, we need to clone the repository
			gitRepository, err = git.PlainClone(repositoryPath, false, &git.CloneOptions{
				URL:           repositoryURL,
				Progress:      os.Stdout,
				Auth:          auth,
				ReferenceName: "refs/heads/testing",
				SingleBranch:  true,
			})
			if err != nil {
				return err
			}
		} else {
			panic(err)
		}

		head, err := gitRepository.Head()
		if err != nil {
			return err
		}

		var previousCommit plumbing.Hash
		previousCommitBytes, err := s.syncedDB.Get(aliasBytes)
		if err != nil && err != database.ErrNotFound {
			return err
		}
		copy(previousCommit[:], previousCommitBytes)

		// Our head should have the latest changes now
		head, err = gitRepository.Head()
		if err != nil {
			return err
		}

		// TODO graceful failure, don't block on a single repo failing to sync
		// If our hashes don't match, we need to re-build our binary
		if head.Hash() != previousCommit {
			fmt.Printf("Changes detected. Re-building binaries for %s@%s.\n", repositoryName, head.Hash())
			//TODO execute build script instead.
			build := exec.Command("go", "build", "-o", fmt.Sprintf("%s/%s", s.buildPath, repositoryName), fmt.Sprintf("%s/main", repositoryPath))
			// Need to set working directory to the same directory where go.mod
			// is or go buildDir won't work.
			build.Dir = repositoryPath

			if err := build.Run(); err != nil {
				return err
			}
			pluginMap := map[string]plugin.Plugin{
				constant.Repository: &grpc.PluginRepository{},
			}

			binaryPath := filepath.Join(s.buildPath, repositoryName)
			client := plugin.NewClient(&plugin.ClientConfig{
				HandshakeConfig:  core.HandshakeConfig,
				Plugins:          pluginMap,
				Cmd:              subprocess.New(binaryPath),
				AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
			})

			rpcClient, err := client.Client()
			if err != nil {
				client.Kill()
				return err
			}

			// request the repository
			raw, err := rpcClient.Dispense(constant.Repository)
			if err != nil {
				client.Kill()
				return err
			}

			repository := raw.(avaxPlugin.Repository)
			//repositoryDB := prefixdb.New([]byte(repositoryName), s.db)

			subnets, err := repository.GetSubnets()
			if err != nil {
				client.Kill()
				return err
			}

			fmt.Printf("loaded subnets: %s", subnets)
			// TODO refactor and use defer
			client.Kill()
		}
	}
	return nil
}

func (s *Service) AddRepository(alias string, url string) error {
	return s.repositoriesDB.Put([]byte(alias), []byte(url))
}

func (s *Service) RemoveRepository(alias string) error {
	aliasBytes := []byte(alias)
	repository := prefixdb.New(aliasBytes, s.db)
	itr := repository.NewIterator()

	// Delete all the plugin definitions in the repository
	for itr.Next() {
		if err := repository.Delete(itr.Key()); err != nil {
			return err
		}
	}

	return s.repositoriesDB.Delete(aliasBytes)
}

func (s *Service) ListRepositories() []string {
	repos := make([]string, 0)
	itr := s.repositoriesDB.NewIterator()
	for itr.Next() {
		repos = append(repos, string(itr.Key()))
	}

	return repos
}

func New(config Config) (*Service, error) {
	dbDir := filepath.Join(config.WorkingDir, dbDir)
	db, err := leveldb.New(dbDir, []byte{}, logging.NoLog{})
	if err != nil {
		return nil, err
	}

	repoDB := prefixdb.New(repoPrefix, db)
	syncedDB := prefixdb.New(syncedPrefix, db)

	//initialize codec
	//c := linearcodec.NewDefault()
	//errs := wrappers.Errs{}
	//errs.Add(
	//	c.RegisterType(&grpc.PluginRepository{}),
	//)
	//if errs.Errored() {
	//	return nil, errs.Err
	//}

	//codecManager := codec.NewDefaultManager()
	//if err := codecManager.RegisterCodec(codecVersion, c); err != nil {
	//	return nil, err
	//}

	s := &Service{
		//codecManager:     codecManager,
		repositoriesPath: filepath.Join(config.WorkingDir, repositoriesDir),
		buildPath:        filepath.Join(config.WorkingDir, buildDir),
		db:               db,
		syncedDB:         syncedDB,
		repositoriesDB:   repoDB,
		fsReader:         config.FsReader,
	}

	if err := os.MkdirAll(s.repositoriesPath, perms.ReadWriteExecute); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(s.buildPath, perms.ReadWriteExecute); err != nil {
		return nil, err
	}

	coreKey := []byte(core.Alias)
	if _, err = repoDB.Get(coreKey); err == database.ErrNotFound {
		err := s.AddRepository(core.Alias, core.URL)
		if err != nil {
			return nil, err
		}
	}

	if _, err := syncedDB.Get(coreKey); err == database.ErrNotFound {
		fmt.Println("Bootstrap not detected. Bootstrapping...")
		err := s.Update()
		if err != nil {
			return nil, err
		}

		fmt.Println("Finished bootstrapping.")
	}

	return s, nil
}

type Config struct {
	WorkingDir string
	FsReader   filesystem.Reader
}
