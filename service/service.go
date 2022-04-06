package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/codec/linearcodec"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/leveldb"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/utils/filesystem"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/perms"
	"github.com/ava-labs/avalanchego/utils/subprocess"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/go-git/go-git/v5"
	"github.com/hashicorp/go-plugin"

	avaxPlugin "github.com/ava-labs/avalanche-plugin/plugin"
	"github.com/ava-labs/avalanche-plugins-core/core"
)

var (
	dbPath           = "db"
	repositoriesPath = "repositories"

	repoPrefix      = []byte("repo")
	bootstrapPrefix = []byte("bootstrap")

	bootstrappedKey = []byte("bootstrapped")

	codecVersion uint16 = 1
)

type Service struct {
	repositoriesPath string

	codecManager   codec.Manager
	db             database.Database
	repositoriesDB database.Database
	bootstrapDB    database.Database
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
	files, err := s.fsReader.ReadDir(s.repositoriesPath)
	if err != nil {
		return err
	}

	filesMap := make(map[string]bool)
	for _, file := range files {
		// we only care about git repositories, which are directories
		if !file.IsDir() {
			continue
		}
		filesMap[file.Name()] = true
	}

	itr := s.repositoriesDB.NewIterator()
	for itr.Next() {
		repositoryName := string(itr.Key())
		repositoryURL := string(itr.Value())
		repositoryPath := filepath.Join(s.repositoriesPath, repositoryName)

		if _, ok := filesMap[repositoryName]; ok {
			// already exists, we need to check out the latest changes
			gitRepository, err := git.PlainOpen(repositoryPath)
			if err != nil {
				return err
			}

			// fetch latest changes
			err = gitRepository.Fetch(&git.FetchOptions{
				RemoteName: "origin",
			})
			if err != nil {
				return err
			}
		} else {
			// otherwise, we need to clone the repository
			_, err := git.PlainClone(repositoryPath, false, &git.CloneOptions{
				URL:      repositoryURL,
				Progress: os.Stdout,
			})
			if err != nil {
				return err
			}
		}

		// build the repository binary
		build := subprocess.New(filepath.Join(repositoryPath, "scripts", "build.sh"))
		if err := build.Run(); err != nil {
			return err
		}

		pluginMap := map[string]plugin.Plugin{
			"repository": &avaxPlugin.RPCRepository{},
		}

		binaryPath := filepath.Join(repositoryPath, "build", repositoryName)
		client := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: core.HandshakeConfig,
			Plugins:         pluginMap,
			Cmd:             subprocess.New(binaryPath),
		})

		rpcClient, err := client.Client()
		if err != nil {
			return err
		}

		// request the repository
		raw, err := rpcClient.Dispense("repository")
		if err != nil {
			return err
		}

		repository := raw.(avaxPlugin.Repository)
		//repositoryDB := prefixdb.New([]byte(repositoryName), s.db)

		plugins, err := repository.Plugins()
		if err != nil {
			return err
		}

		fmt.Printf("loaded plugins: %s", plugins)

		//for _, subnet := range plugins.Subnets {
		//	subnetKey := []byte(subnet.Alias())
		//
		//	bytes, err := repositoryDB.Get(subnetKey)
		//	if err != nil && err != database.ErrNotFound {
		//		return err
		//	}
		//	if err == database.ErrNotFound {
		//
		//	}
		//
		//	if err := repositoryDB.Put([]byte(subnet.Alias()), subnet); err != nil {
		//		return err
		//	}
		//}
		// DO this for vms too

		client.Kill()
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
	dbDir := filepath.Join(config.WorkingDir, dbPath)
	db, err := leveldb.New(dbDir, []byte{}, logging.NoLog{})
	if err != nil {
		return nil, err
	}

	repoDB := prefixdb.New(repoPrefix, db)
	bootstrapDB := prefixdb.New(bootstrapPrefix, db)

	// initialize codec
	c := linearcodec.NewDefault()
	errs := wrappers.Errs{}
	errs.Add(
		c.RegisterType(&avaxPlugin.RPCRepository{}),
	)
	if errs.Errored() {
		return nil, errs.Err
	}

	codecManager := codec.NewDefaultManager()
	if err := codecManager.RegisterCodec(codecVersion, c); err != nil {
		return nil, err
	}

	s := &Service{
		codecManager:     codecManager,
		repositoriesPath: filepath.Join(config.WorkingDir, repositoriesPath),
		db:               db,
		repositoriesDB:   repoDB,
		fsReader:         config.FsReader,
	}

	if err := os.MkdirAll(s.repositoriesPath, perms.ReadWriteExecute); err != nil {
		return nil, err
	}

	if _, err = repoDB.Get([]byte(core.Name)); err == database.ErrNotFound {
		err := s.AddRepository(core.Name, core.URL)
		if err != nil {
			return nil, err
		}
	}

	if _, err := bootstrapDB.Get(bootstrappedKey); err == database.ErrNotFound {
		err := s.Update()
		if err != nil {
			return nil, err
		}

		if err := bootstrapDB.Put(bootstrappedKey, bootstrappedKey); err != nil {
			return nil, err
		}
	}

	return s, nil
}

type Config struct {
	WorkingDir string
	FsReader   filesystem.Reader
}
