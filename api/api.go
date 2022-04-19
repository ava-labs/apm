package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

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
	dbDir        = "db"
	repositories = "repositories"
	subnets      = "subnets"
	vms          = "vms"

	repoPrefix   = []byte("repo")
	vmPrefix     = []byte("vm")
	subnetPrefix = []byte("subnet")

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

	codecManager codec.Manager

	db           database.Database
	repositoryDB database.Database
	subnetDB     database.Database
	vmDB         database.Database
}

func (s *Service) Install(alias string) error {
	var vm = &types.VM{}

	itr := s.repositoryDB.NewIterator()

	for itr.Next() {
		repoDB := prefixdb.New(itr.Key(), s.db)
		vmDB := prefixdb.New(vmPrefix, repoDB)

		vmBytes, err := vmDB.Get([]byte(alias))
		if err == database.ErrNotFound {
			// This alias didn't exist in this repository
			continue
		}
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(vmBytes, vm)
		if err != nil {
			return err
		}
	}

	if vm.URL != "" {

	}

	if vm.InstallScript != "" {
		cmd := exec.Cmd{
			Path: vm.InstallScript,
			Dir:  "",
		}
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) Uninstall(alias string) error {
	return nil
}

func (s *Service) Join(alias string) error {
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
	itr := s.repositoryDB.NewIterator()

	globalVMs := prefixdb.New(vmPrefix, s.db)
	globalSubnets := prefixdb.New(subnetPrefix, s.db)

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
			return err
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

		repoDB := prefixdb.New(aliasBytes, s.repositoryDB)

		vmsPath := filepath.Join(repositoryPath, vms)
		repoVMs := prefixdb.New(vmPrefix, repoDB)
		if err := loadFromYAML[*types.VM](vmKey, vmsPath, aliasBytes, latestCommit, globalVMs, repoVMs); err != nil {
			return err
		}

		subnetsPath := filepath.Join(repositoryPath, subnets)
		repoSubnets := prefixdb.New(subnetPrefix, repoDB)
		if err := loadFromYAML[*types.Subnet](subnetKey, subnetsPath, aliasBytes, latestCommit, globalSubnets, repoSubnets); err != nil {
			return err
		}

		// Now we need to delete anything that wasn't updated in the latest commit
		if err := deleteStalePlugins[*types.VM](repoVMs, latestCommit); err != nil {
			return err
		}
		if err := deleteStalePlugins[*types.Subnet](repoSubnets, latestCommit); err != nil {
			return err
		}

		updatedMetadata := repository.Metadata{
			Alias:  repositoryMetadata.Alias,
			URL:    repositoryMetadata.URL,
			Commit: latestCommit,
		}
		updatedMetadataBytes, err := yaml.Marshal(updatedMetadata)
		if err != nil {
			return err
		}

		if err := s.repositoryDB.Put(aliasBytes, updatedMetadataBytes); err != nil {
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

func deleteStalePlugins[T types.Plugin](db database.Database, latestCommit plumbing.Hash) error {
	itr := db.NewIterator()
	batch := db.NewBatch()

	for itr.Next() {
		record := &repository.Record[T]{}
		if err := yaml.Unmarshal(itr.Value(), record); err != nil {
			return nil
		}

		if record.Commit != latestCommit {
			fmt.Printf("Deleting a stale plugin: %s@%s as of %s.\n", record.Plugin.Alias(), record.Commit, latestCommit)
			if err := batch.Delete(itr.Key()); err != nil {
				return err
			}
		}
	}

	if err := batch.Write(); err != nil {
		return err
	}
	return nil
}

func (s *Service) AddRepository(alias string, url string) error {
	//TODO should be idempotent
	metadata := repository.Metadata{
		Alias:  alias,
		URL:    url,
		Commit: plumbing.ZeroHash, // hasn't been synced yet
	}
	metadataBytes, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}
	return s.repositoryDB.Put([]byte(alias), metadataBytes)
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
	//TODO remove from subnets + vms

	// remove it from our list of tracked repositories
	return s.repositoryDB.Delete(aliasBytes)
}

func (s *Service) ListRepositories() error {
	itr := s.repositoryDB.NewIterator()

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "alias\turl")
	for itr.Next() {
		metadata := &repository.Metadata{}
		if err := yaml.Unmarshal(itr.Value(), metadata); err != nil {
			return err
		}

		fmt.Fprintln(w, fmt.Sprintf("%s\t%s", metadata.Alias, metadata.URL))
	}
	w.Flush()
	return nil
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
		repositoriesPath: filepath.Join(config.WorkingDir, repositories),
		db:               db,
		repositoryDB:     repoDB,
	}

	if err := os.MkdirAll(s.repositoriesPath, perms.ReadWriteExecute); err != nil {
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
	repositoryMetadataBytes, err := s.repositoryDB.Get(alias)
	if err != nil && err != database.ErrNotFound {
		return nil, err
	}

	repositoryMetadata := &repository.Metadata{}
	if err := yaml.Unmarshal(repositoryMetadataBytes, repositoryMetadata); err != nil {
		return nil, err
	}

	return repositoryMetadata, nil
}

type Config struct {
	WorkingDir string
}
