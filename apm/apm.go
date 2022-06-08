package apm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/ava-labs/avalanche-plugins-core/core"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/leveldb"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/perms"
	"github.com/ava-labs/avalanchego/version"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/afero"

	"github.com/ava-labs/apm/admin"
	"github.com/ava-labs/apm/engine"
	"github.com/ava-labs/apm/git"
	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/types"
	"github.com/ava-labs/apm/url"
	"github.com/ava-labs/apm/util"
	"github.com/ava-labs/apm/workflow"
)

var (
	dbDir         = "db"
	repositoryDir = "repositories"
	tmpDir        = "tmp"
)

type Config struct {
	Directory        string
	Auth             http.BasicAuth
	AdminApiEndpoint string
	PluginDir        string
	Fs               afero.Fs
}

type APM struct {
	db database.Database

	sourcesList  storage.Storage[storage.SourceInfo]
	installedVMs storage.Storage[version.Semantic]
	registry     storage.Storage[storage.RepoList]
	repoFactory  storage.RepositoryFactory

	engine workflow.Executor

	auth http.BasicAuth

	adminClient admin.Client
	installer   workflow.Installer

	repositoriesPath string
	tmpPath          string
	pluginPath       string
	fs               afero.Fs
}

func New(config Config) (*APM, error) {
	dbDir := filepath.Join(config.Directory, dbDir)
	db, err := leveldb.New(dbDir, []byte{}, logging.NoLog{})
	if err != nil {
		return nil, err
	}

	var a = &APM{
		repositoriesPath: filepath.Join(config.Directory, repositoryDir),
		tmpPath:          filepath.Join(config.Directory, tmpDir),
		pluginPath:       config.PluginDir,
		db:               db,
		registry:         storage.NewRegistry(db),
		sourcesList:      storage.NewSourceInfo(db),
		installedVMs:     storage.NewInstalledVMs(db),
		auth:             config.Auth,
		adminClient: admin.NewHttpClient(
			admin.HttpClientConfig{
				Endpoint: fmt.Sprintf("http://%s", config.AdminApiEndpoint),
			},
		),
		installer: workflow.NewVMInstaller(
			workflow.VMInstallerConfig{
				Fs:        config.Fs,
				UrlClient: url.NewHttpClient(),
			},
		),
		engine:      engine.NewWorkflowEngine(),
		fs:          config.Fs,
		repoFactory: storage.NewRepositoryFactory(db),
	}
	if err := os.MkdirAll(a.repositoriesPath, perms.ReadWriteExecute); err != nil {
		return nil, err
	}

	// TODO simplify this
	coreKey := []byte(core.Alias)
	if ok, err := a.sourcesList.Has(coreKey); err != nil {
		return nil, err
	} else if !ok {
		err := a.AddRepository(core.Alias, core.URL)
		if err != nil {
			return nil, err
		}
	}

	repoMetadata, err := a.sourcesList.Get(coreKey)
	if err != nil {
		return nil, err
	}

	if repoMetadata.Commit == plumbing.ZeroHash {
		fmt.Println("Bootstrap not detected. Bootstrapping...")
		err := a.Update()
		if err != nil {
			return nil, err
		}

		fmt.Println("Finished bootstrapping.")
	}
	return a, nil
}

func parseAndRun(alias string, registry storage.Storage[storage.RepoList], command func(string) error) error {
	if qualifiedName(alias) {
		return command(alias)
	}

	fullName, err := getFullNameForAlias(registry, alias)
	if err != nil {
		return err
	}

	return command(fullName)

}

func (a *APM) Install(alias string) error {
	return parseAndRun(alias, a.registry, a.install)
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

	repoAlias, plugin := util.ParseQualifiedName(name)
	organization, repo := util.ParseAlias(repoAlias)

	repository := a.repoFactory.GetRepository([]byte(repoAlias))

	workflow := workflow.NewInstall(workflow.InstallConfig{
		Name:         name,
		Plugin:       plugin,
		Organization: organization,
		Repo:         repo,
		TmpPath:      a.tmpPath,
		PluginPath:   a.pluginPath,
		InstalledVMs: a.installedVMs,
		VMStorage:    repository.VMs,
		Fs:           a.fs,
		Installer:    a.installer,
	})

	return a.engine.Execute(workflow)
}

func (a *APM) Uninstall(alias string) error {
	return parseAndRun(alias, a.registry, a.uninstall)
}

func (a *APM) uninstall(name string) error {
	alias, plugin := util.ParseQualifiedName(name)

	repository := a.repoFactory.GetRepository([]byte(alias))

	wf := workflow.NewUninstall(
		workflow.UninstallConfig{
			Name:         name,
			Plugin:       plugin,
			RepoAlias:    alias,
			VMStorage:    repository.VMs,
			InstalledVMs: a.installedVMs,
		},
	)

	return wf.Execute()
}

func (a *APM) JoinSubnet(alias string) error {
	return parseAndRun(alias, a.registry, a.joinSubnet)
}

func (a *APM) joinSubnet(fullName string) error {
	alias, plugin := util.ParseQualifiedName(fullName)
	repoRegistry := a.repoFactory.GetRepository([]byte(alias))

	var (
		// weird hack for generics
		definition storage.Definition[types.Subnet]
		err        error
	)

	definition, err = repoRegistry.Subnets.Get([]byte(plugin))
	if err != nil {
		return err
	}

	// definition := &storage.Definition[types.Subnet]{}
	// if err := yaml.Unmarshal(subnetBytes, definition); err != nil {
	//	return err
	// }
	subnet := definition.Definition

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

	fullName, err := getFullNameForAlias(a.registry, alias)
	if err != nil {
		return err
	}

	return a.info(fullName)
}

func (a *APM) info(fullName string) error {
	return nil
}

func (a *APM) Update() error {
	workflow := workflow.NewUpdate(workflow.UpdateConfig{
		Executor:         a.engine,
		Registry:         a.registry,
		InstalledVMs:     a.installedVMs,
		DB:               a.db,
		TmpPath:          a.tmpPath,
		PluginPath:       a.pluginPath,
		Installer:        a.installer,
		SourcesList:      a.sourcesList,
		RepositoriesPath: a.repositoriesPath,
		Auth:             a.auth,
		GitFactory:       git.RepositoryFactory{},
	})

	if err := a.engine.Execute(workflow); err != nil {
		return err
	}

	return nil
}

func (a *APM) AddRepository(alias string, url string) error {
	wf := workflow.NewAddRepository(
		workflow.AddRepositoryConfig{
			SourcesList: a.sourcesList,
			Alias:       alias,
			Url:         url,
		},
	)

	return a.engine.Execute(wf)
}

func (a *APM) RemoveRepository(alias string) error {
	if qualifiedName(alias) {
		return a.removeRepository(alias)
	}

	fullName, err := getFullNameForAlias(a.registry, alias)
	if err != nil {
		return err
	}

	return a.removeRepository(fullName)
}

func (a *APM) removeRepository(name string) error {
	if name == core.Alias {
		fmt.Printf("Can't remove %s (required repository).\n", core.Alias)
		return nil
	}

	// TODO don't let people remove core
	aliasBytes := []byte(name)
	repoRegistry := a.repoFactory.GetRepository(aliasBytes)

	// delete all the plugin definitions in the repository
	vmItr := repoRegistry.VMs.Iterator()
	defer vmItr.Release()

	for vmItr.Next() {
		if err := repoRegistry.VMs.Delete(vmItr.Key()); err != nil {
			return err
		}
	}

	subnetItr := repoRegistry.Subnets.Iterator()
	defer subnetItr.Release()

	for subnetItr.Next() {
		if err := repoRegistry.Subnets.Delete(subnetItr.Key()); err != nil {
			return err
		}
	}

	// remove it from our list of tracked repositories
	return a.sourcesList.Delete(aliasBytes)
}

func (a *APM) ListRepositories() error {
	itr := a.sourcesList.Iterator()

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "alias\turl")
	for itr.Next() {
		metadata, err := itr.Value()
		if err != nil {
			return err
		}

		fmt.Fprintln(w, fmt.Sprintf("%s\t%s", metadata.Alias, metadata.URL))
	}
	w.Flush()
	return nil
}

func qualifiedName(name string) bool {
	parsed := strings.Split(name, ":")
	return len(parsed) > 1
}

func getFullNameForAlias(registry storage.Storage[storage.RepoList], alias string) (string, error) {
	repoList, err := registry.Get([]byte(alias))
	if err != nil {
		return "", err
	}

	if len(repoList.Repositories) > 1 {
		return "", errors.New(fmt.Sprintf("more than one match found for %s. Please specify the fully qualified name. Matches: %s.\n", alias, repoList.Repositories))
	}

	return fmt.Sprintf("%s:%s", repoList.Repositories[0], alias), nil
}
