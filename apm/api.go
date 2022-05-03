package apm

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/leveldb"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/perms"
	"github.com/cavaliergopher/grab/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/yaml.v2"

	"github.com/ava-labs/avalanche-plugins-core/core"

	"github.com/ava-labs/apm/admin"
	"github.com/ava-labs/apm/repository"
	"github.com/ava-labs/apm/types"
)

var (
	dbDir        = "db"
	repositories = "repositories"
	tmp          = "tmp"
	subnets      = "subnets"
	vms          = "vms"

	repoPrefix         = []byte("repo")
	vmPrefix           = []byte("vm")
	subnetPrefix       = []byte("subnet")
	installedVMsPrefix = []byte("installed_vms")

	vmKey     = "vm"
	subnetKey = "subnet"
)

type Config struct {
	Directory        string
	Auth             gitHttp.BasicAuth
	AdminApiEndpoint string
	PluginDir        string
}

type APM struct {
	repositoriesPath string
	tmpPath          string
	pluginPath       string

	db           database.Database
	repositoryDB database.Database
	subnetDB     database.Database
	vmDB         database.Database

	installedVMs database.Database

	auth gitHttp.BasicAuth

	adminClient admin.Client
}

func New(config Config) (*APM, error) {
	dbDir := filepath.Join(config.Directory, dbDir)
	db, err := leveldb.New(dbDir, []byte{}, logging.NoLog{})
	if err != nil {
		return nil, err
	}

	s := &APM{
		repositoriesPath: filepath.Join(config.Directory, repositories),
		tmpPath:          filepath.Join(config.Directory, tmp),
		pluginPath:       config.PluginDir,
		db:               db,
		repositoryDB:     prefixdb.New(repoPrefix, db),
		subnetDB:         prefixdb.New(subnetPrefix, db),
		vmDB:             prefixdb.New(vmPrefix, db),
		installedVMs:     prefixdb.New(installedVMsPrefix, db),
		auth:             config.Auth,
		adminClient: admin.NewHttpClient(
			admin.HttpClientConfig{
				Endpoint: fmt.Sprintf("http://%s", config.AdminApiEndpoint),
			},
		),
	}

	if err := os.MkdirAll(s.repositoriesPath, perms.ReadWriteExecute); err != nil {
		return nil, err
	}

	//TODO simplify this
	coreKey := []byte(core.Alias)
	if _, err = s.repositoryDB.Get(coreKey); err == database.ErrNotFound {
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

func (a *APM) Install(alias string) error {
	if qualifiedName(alias) {
		return a.install(alias)
	}

	fullName, err := getFullNameForAlias(a.vmDB, alias)
	if err != nil {
		return err
	}

	return a.install(fullName)
}

func (a *APM) install(name string) error {
	nameBytes := []byte(name)

	ok, err := a.installedVMs.Has(nameBytes)
	if err != nil {
		return err
	}

	if ok {
		fmt.Printf("VM %s is already installed. Skipping.\n", name)
		return nil
	}

	alias, plugin := parseQualifiedName(name)
	organization, repo := parseAlias(alias)

	repoDB := prefixdb.New([]byte(alias), a.db)
	repoVMDB := prefixdb.New(vmPrefix, repoDB)

	bytes, err := repoVMDB.Get([]byte(plugin))
	if err != nil {
		return err
	}

	record := &repository.Plugin[*types.VM]{}
	if err := yaml.Unmarshal(bytes, record); err != nil {
		return err
	}

	vm := record.Plugin
	archiveFile := fmt.Sprintf("%s.tar.gz", plugin)
	tmpPath := filepath.Join(a.tmpPath, organization, repo)
	downloadPath := filepath.Join(tmpPath, archiveFile)

	if vm.InstallScript == "" {
		fmt.Printf("No install script found for %s.", name)
		return nil
	}

	// Download the .tar.gz file from the url
	client := grab.NewClient()
	req, _ := grab.NewRequest(downloadPath, vm.URL)

	// Start downloading
	fmt.Printf("Downloading %v...\n", req.URL())
	resp := client.Do(req)
	fmt.Printf("HTTP response %v\n", resp.HTTPResponse.Status)

	// Start progress loop
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		return errors.New(fmt.Sprintf("Download failed: %s", err))
	}

	// Create the directory we'll store the plugin sources in if it doesn't exist.
	if _, err := os.Stat(plugin); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Creating sources directory...\n")
		cmd := exec.Command("mkdir", plugin)
		cmd.Dir = tmpPath

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	fmt.Printf("Uncompressing %s...\n", name)
	cmd := exec.Command("tar", "xf", archiveFile, "-C", plugin, "--strip-components", "1")
	cmd.Dir = tmpPath
	if err := cmd.Run(); err != nil {
		return err
	}

	workingDir := filepath.Join(tmpPath, plugin)
	fmt.Printf("Running install script at %s...\n", vm.InstallScript)
	cmd = exec.Command(vm.InstallScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = workingDir
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Printf("Moving binary %s into plugin directory...\n", vm.ID_)
	if err := os.Rename(filepath.Join(workingDir, vm.BinaryPath), filepath.Join(a.pluginPath, vm.ID_)); err != nil {
		panic(err)
		return err
	}

	fmt.Printf("Cleaning up temporary files...\n")
	if err := os.Remove(filepath.Join(tmpPath, archiveFile)); err != nil {
		panic(err)
		return err
	}

	if err := os.RemoveAll(filepath.Join(tmpPath, plugin)); err != nil {
		panic(err)
		return err
	}

	fmt.Printf("Adding virtual machine %s to installation registry...\n", vm.ID_)
	installedVersion, err := yaml.Marshal(vm.Version)
	if err != nil {
		return err
	}
	if err := a.installedVMs.Put(nameBytes, installedVersion); err != nil {
		return err
	}

	fmt.Printf("Successfully installed %s@v%v.%v.%v.\n", name, vm.Version.Major(), vm.Version.Minor(), vm.Version.Patch())
	return nil
}

func (a *APM) Uninstall(alias string) error {
	if qualifiedName(alias) {
		return a.uninstall(alias)
	}

	fullName, err := getFullNameForAlias(a.vmDB, alias)
	if err != nil {
		return err
	}

	return a.uninstall(fullName)
}

func (a *APM) uninstall(name string) error {
	nameBytes := []byte(name)

	ok, err := a.installedVMs.Has(nameBytes)
	if err != nil {
		return err
	}

	if !ok {
		fmt.Printf("VM %s is already not installed. Skipping.\n", name)
		return nil
	}

	alias, plugin := parseQualifiedName(name)

	repoDB := prefixdb.New([]byte(alias), a.db)
	repoVMDB := prefixdb.New(vmPrefix, repoDB)

	ok, err = repoVMDB.Has([]byte(plugin))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Printf("Virtual machine already %s doesn't exist in the vm registry for %s.", name, alias)
		return nil
	}

	if err := a.installedVMs.Delete([]byte(plugin)); err != nil {
		return err
	}

	fmt.Printf("Successfully uninstalled %s.", name)

	return nil
}

func (a *APM) JoinSubnet(alias string) error {
	if qualifiedName(alias) {
		return a.joinSubnet(alias)
	}

	fullName, err := getFullNameForAlias(a.subnetDB, alias)
	if err != nil {
		return err
	}

	return a.joinSubnet(fullName)
}

func (a *APM) joinSubnet(fullName string) error {
	alias, plugin := parseQualifiedName(fullName)
	aliasBytes := []byte(alias)
	repoDB := prefixdb.New(aliasBytes, a.db)
	repoSubnetDB := prefixdb.New(subnetPrefix, repoDB)

	subnetBytes, err := repoSubnetDB.Get([]byte(plugin))
	if err != nil {
		return err
	}

	record := &repository.Plugin[*types.Subnet]{}
	if err := yaml.Unmarshal(subnetBytes, record); err != nil {
		return err
	}
	subnet := record.Plugin

	// TODO prompt user, add force flag
	fmt.Printf("Installing virtual machines for subnet %s.\n", subnet.ID())
	for _, vm := range subnet.VMs_ {
		if err := a.Install(vm); err != nil {
			return err
		}
	}

	fmt.Printf("Updating virtual machines...\n")
	if err := a.adminClient.LoadVMs(); err != nil {
		return err
	}

	fmt.Printf("Whitelisting subnet %s...\n", subnet.ID())
	if err := a.adminClient.WhitelistSubnet(subnet.ID()); err != nil {
		return err
	}

	fmt.Printf("Finished installing virtual machines for subnet %s.\n", subnet.ID_)
	return nil
}

func (a *APM) Upgrade(alias string) error {
	return nil
}

func (a *APM) Search(alias string) error {
	return nil
}

func (a *APM) Info(alias string) error {
	if qualifiedName(alias) {
		return a.install(alias)
	}

	fullName, err := getFullNameForAlias(a.vmDB, alias)
	if err != nil {
		return err
	}

	return a.info(fullName)
}

func (a *APM) info(fullName string) error {
	return nil
}

func (a *APM) Update() error {
	itr := a.repositoryDB.NewIterator()

	for itr.Next() {
		// Need to split the alias to support Windows
		aliasBytes := itr.Key()
		organization, repo := parseAlias(string(aliasBytes))

		repositoryMetadata, err := a.repositoryMetadataFor(aliasBytes)
		repositoryPath := filepath.Join(a.repositoriesPath, organization, repo)
		gitRepository, err := a.syncRepository(repositoryMetadata.URL, repositoryPath, "refs/heads/main")

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
			fmt.Printf("Already at latest for %s@%s.\n", repo, previousCommit)
			continue
		}

		repoDB := prefixdb.New(aliasBytes, a.db)

		vmsPath := filepath.Join(repositoryPath, vms)
		repoVMDB := prefixdb.New(vmPrefix, repoDB)
		if err := loadFromYAML[*types.VM](vmKey, vmsPath, aliasBytes, latestCommit, a.vmDB, repoVMDB); err != nil {
			return err
		}

		subnetsPath := filepath.Join(repositoryPath, subnets)
		repoSubnetDB := prefixdb.New(subnetPrefix, repoDB)
		if err := loadFromYAML[*types.Subnet](subnetKey, subnetsPath, aliasBytes, latestCommit, a.subnetDB, repoSubnetDB); err != nil {
			return err
		}

		// Now we need to delete anything that wasn't updated in the latest commit
		if err := deleteStalePlugins[*types.VM](repoVMDB, latestCommit); err != nil {
			return err
		}
		if err := deleteStalePlugins[*types.Subnet](repoSubnetDB, latestCommit); err != nil {
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

		if err := a.repositoryDB.Put(aliasBytes, updatedMetadataBytes); err != nil {
			return err
		}

		if previousCommit == plumbing.ZeroHash {
			fmt.Printf("Finished initializing %s@%s.\n", repo, latestCommit)
		} else {
			fmt.Printf("Finished updating from %s to %s@%s.\n", previousCommit, repo, latestCommit)
		}
	}

	return nil
}

func deleteStalePlugins[T types.Plugin](db database.Database, latestCommit plumbing.Hash) error {
	itr := db.NewIterator()
	batch := db.NewBatch()

	for itr.Next() {
		record := &repository.Plugin[T]{}
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

func (a *APM) AddRepository(alias string, url string) error {
	//TODO don't let people remove core
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
	return a.repositoryDB.Put([]byte(alias), metadataBytes)
}

func (a *APM) RemoveRepository(alias string) error {
	aliasBytes := []byte(alias)
	repoDB := prefixdb.New(aliasBytes, a.db)
	itr := repoDB.NewIterator()

	// delete all the plugin definitions in the repository
	for itr.Next() {
		if err := repoDB.Delete(itr.Key()); err != nil {
			return err
		}
	}
	//TODO remove from subnets + vms

	// remove it from our list of tracked repositories
	return a.repositoryDB.Delete(aliasBytes)
}

func (a *APM) ListRepositories() error {
	itr := a.repositoryDB.NewIterator()

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

func (a *APM) repositoryMetadataFor(alias []byte) (*repository.Metadata, error) {
	repositoryMetadataBytes, err := a.repositoryDB.Get(alias)
	if err != nil && err != database.ErrNotFound {
		return nil, err
	}

	repositoryMetadata := &repository.Metadata{}
	if err := yaml.Unmarshal(repositoryMetadataBytes, repositoryMetadata); err != nil {
		return nil, err
	}

	return repositoryMetadata, nil
}

func (a *APM) syncRepository(url string, path string, reference plumbing.ReferenceName) (*git.Repository, error) {
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
				Auth:          &a.auth,
				Progress:      io.Discard,
			},
		)
	} else if os.IsNotExist(err) {
		// otherwise, we need to clone the repository
		gitRepository, err = git.PlainClone(path, false, &git.CloneOptions{
			URL:           url,
			ReferenceName: reference,
			SingleBranch:  true,
			Auth:          &a.auth,
			Progress:      io.Discard,
		})
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return gitRepository, nil
}

func qualifiedName(name string) bool {
	parsed := strings.Split(name, ":")
	return len(parsed) > 1
}

func getFullNameForAlias(db database.Database, alias string) (string, error) {
	bytes, err := db.Get([]byte(alias))
	if err != nil {
		return "", err
	}

	registry := &repository.Registry{}
	if err := yaml.Unmarshal(bytes, registry); err != nil {
		return "", err
	}

	if len(registry.Repositories) > 1 {
		return "", errors.New(fmt.Sprintf("more than one match found for %s. Please specify the fully qualified name. Matches: %s.\n", alias, registry.Repositories))
	}

	return fmt.Sprintf("%s:%s", registry.Repositories[0], alias), nil
}
