package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/codec/linearcodec"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/leveldb"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/perms"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/yaml.v2"

	"github.com/ava-labs/apm/repository"
	"github.com/ava-labs/apm/types"
	"github.com/ava-labs/avalanche-plugins-core/core"
)

var (
	dbDir           = "db"
	repositoriesDir = "repositories"
	subnetsDir      = "subnets"
	vmsDir          = "vms"
	buildDir        = "build"

	repoPrefix   = []byte("repo")
	syncedPrefix = []byte("synced")

	vmKey     = "vm"
	subnetKey = "subnet"

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

	codecManager   codec.Manager
	db             database.Database
	repositoriesDB database.Database
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

		repositoryMetadata, err := s.repositoryMetadataFor(aliasBytes)
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
					ReferenceName: "refs/heads/yaml",
					SingleBranch:  true,
					Auth:          auth,
					Progress:      ioutil.Discard,
				},
			)
		} else if os.IsNotExist(err) {
			// otherwise, we need to clone the repository
			gitRepository, err = git.PlainClone(repositoryPath, false, &git.CloneOptions{
				URL:           repositoryMetadata.URL,
				ReferenceName: "refs/heads/yaml",
				SingleBranch:  true,
				Auth:          auth,
				Progress:      ioutil.Discard,
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

		previousCommit := repositoryMetadata.Commit

		// Our head should have the latest changes now
		head, err = gitRepository.Head()
		if err != nil {
			return err
		}
		latestCommit := head.Hash()

		if latestCommit == previousCommit {
			fmt.Printf("Already at latest for %s@%s.\n", repositoryName, previousCommit)
			continue
		}

		vmsPath := filepath.Join(repositoryPath, vmsDir)
		if err := loadFromYAML[*types.VM](vmKey, vmsPath, aliasBytes, s.db); err != nil {
			return err
		}

		subnetsPath := filepath.Join(repositoryPath, subnetsDir)
		if err := loadFromYAML[*types.Subnet](subnetKey, subnetsPath, aliasBytes, s.db); err != nil {
			return err
		}
		updatedMetadata := repository.Metadata{
			Alias:  repositoryMetadata.Alias,
			URL:    repositoryMetadata.URL,
			Commit: latestCommit,
		}
		updatedMetadataBytes, err := json.Marshal(updatedMetadata)

		if err != nil {
			return err
		}

		if err := s.repositoriesDB.Put(aliasBytes, updatedMetadataBytes); err != nil {
			return err
		}

		if previousCommit == plumbing.ZeroHash {
			fmt.Printf("Finished initializing %s@%s.\n", repositoryName, latestCommit)
		} else {
			fmt.Printf("Finished updating from %s to %s@%s.\n", previousCommit, repositoryName, latestCommit)
		}
	}

	return nil
}

func loadFromYAML[T types.Plugin](
	key string,
	path string,
	repositoryAlias []byte,
	db database.Database,
) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	repositoryDB := prefixdb.New(repositoryAlias, db)
	batch := repositoryDB.NewBatch()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		nameWithExtension := file.Name()
		// Strip any extension from the file. This is to support windows .exe
		// files.
		name := nameWithExtension[:len(nameWithExtension)-len(filepath.Ext(nameWithExtension))]

		// Skip hidden files.
		if len(name) == 0 {
			continue
		}

		bytes, err := os.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return err
		}
		data := make(map[string]T)

		if err := yaml.Unmarshal(bytes, data); err != nil {
			return err
		}

		if err := batch.Put([]byte(data[key].Alias()), []byte(file.Name())); err != nil {
			return err
		}
	}
	if err := batch.Write(); err != nil {
		return err
	}

	return nil
}

func (s *Service) AddRepository(alias string, url string) error {
	metadata := repository.Metadata{
		Alias:  alias,
		URL:    url,
		Commit: plumbing.ZeroHash,
	}
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	return s.repositoriesDB.Put([]byte(alias), metadataBytes)
}

func (s *Service) RemoveRepository(alias string) error {
	aliasBytes := []byte(alias)
	repoDB := prefixdb.New(aliasBytes, s.db)
	itr := repoDB.NewIterator()

	// delete all the plugin definitions in the repository
	for itr.Next() {
		if err := repoDB.Delete(itr.Key()); err != nil {
			return err
		}
	}

	// remove it from our list of tracked repositories
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

	//initialize codec
	c := linearcodec.NewDefault()
	errs := wrappers.Errs{}
	//errs.Add(
	//	c.RegisterType(&grpc.Subnet{}),
	//	c.RegisterType(&grpc.VM{}),
	//)
	if errs.Errored() {
		return nil, errs.Err
	}

	codecManager := codec.NewDefaultManager()
	if err := codecManager.RegisterCodec(codecVersion, c); err != nil {
		return nil, err
	}

	s := &Service{
		codecManager:     codecManager,
		repositoriesPath: filepath.Join(config.WorkingDir, repositoriesDir),
		buildPath:        filepath.Join(config.WorkingDir, buildDir),
		db:               db,
		repositoriesDB:   repoDB,
	}

	if err := os.MkdirAll(s.repositoriesPath, perms.ReadWriteExecute); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(s.buildPath, perms.ReadWriteExecute); err != nil {
		return nil, err
	}

	//TODO simplify this
	coreKey := []byte(core.Alias)
	if _, err = repoDB.Get(coreKey); err == database.ErrNotFound {
		err := s.AddRepository(core.Alias, core.URL)
		if err != nil {
			return nil, err
		}
	}

	repoMetadata, err := s.repositoryMetadataFor(coreKey)
	if err != nil {
		return nil, err
	}

	if repoMetadata.Commit == plumbing.ZeroHash {
		fmt.Println("Bootstrap not detected. Bootstrapping...")
		err := s.Update()
		if err != nil {
			return nil, err
		}

		fmt.Println("Finished bootstrapping.")
	}
	return s, nil
}

func (s *Service) repositoryMetadataFor(alias []byte) (*repository.Metadata, error) {
	repositoryMetadataBytes, err := s.repositoriesDB.Get(alias)
	if err != nil && err != database.ErrNotFound {
		return nil, err
	}

	repositoryMetadata := &repository.Metadata{}
	if err := json.Unmarshal(repositoryMetadataBytes, repositoryMetadata); err != nil {
		return nil, err
	}

	return repositoryMetadata, nil
}

type Config struct {
	WorkingDir string
}
