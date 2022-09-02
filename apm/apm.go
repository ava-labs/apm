// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package apm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/ava-labs/avalanchego/utils/perms"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/juju/fslock"
	"github.com/spf13/afero"

	"github.com/ava-labs/apm/admin"
	"github.com/ava-labs/apm/constant"
	"github.com/ava-labs/apm/engine"
	"github.com/ava-labs/apm/git"
	"github.com/ava-labs/apm/state"
	"github.com/ava-labs/apm/url"
	"github.com/ava-labs/apm/util"
	"github.com/ava-labs/apm/workflow"
)

const (
	repositoryDir = "repositories"
	tmpDir        = "tmp"
	lockFile      = "apm.lock"
)

type Config struct {
	Directory        string
	Auth             http.BasicAuth
	AdminAPIEndpoint string
	PluginDir        string
	Fs               afero.Fs
	StateFile        state.File
}

type APM struct {
	repoFactory state.RepositoryFactory
	git         git.Factory

	executor workflow.Executor

	auth http.BasicAuth

	adminClient admin.Client
	installer   workflow.Installer

	repositoriesPath string
	tmpPath          string
	pluginPath       string
	adminAPIEndpoint string
	fs               afero.Fs
	stateFile        state.File
	lock             *fslock.Lock
}

func New(config Config) (*APM, error) {
	if err := os.MkdirAll(config.Directory, perms.ReadWriteExecute); err != nil {
		return nil, err
	}
	stateFile, err := state.New(config.Directory)
	if err != nil {
		return nil, err
	}

	repositoriesPath := filepath.Join(config.Directory, repositoryDir)
	a := &APM{
		repoFactory: state.NewRepositoryFactory(repositoriesPath),
		git:         git.RepositoryFactory{},
		executor:    engine.NewWorkflowEngine(stateFile),
		auth:        config.Auth,
		adminClient: admin.NewClient(fmt.Sprintf("http://%s", config.AdminAPIEndpoint)),
		installer: workflow.NewVMInstaller(
			workflow.VMInstallerConfig{
				Fs:        config.Fs,
				URLClient: url.NewClient(),
			},
		),
		repositoriesPath: repositoriesPath,
		tmpPath:          filepath.Join(config.Directory, tmpDir),
		pluginPath:       config.PluginDir,
		adminAPIEndpoint: config.AdminAPIEndpoint,
		fs:               config.Fs,
		stateFile:        stateFile,
		lock:             fslock.New(filepath.Join(config.Directory, lockFile)),
	}
	if err := os.MkdirAll(a.repositoriesPath, perms.ReadWriteExecute); err != nil {
		return nil, err
	}

	// Sync the core repository if it hasn't been bootstrapped yet.
	if _, ok := a.stateFile.Sources[constant.CoreAlias]; !ok {
		err := a.AddRepository(constant.CoreAlias, constant.CoreURL, constant.CoreBranch)
		if err != nil {
			return nil, err
		}
	}

	// Guaranteed to have this now since we've bootstrapped
	repoMetadata := a.stateFile.Sources[constant.CoreAlias]

	if repoMetadata.Commit == plumbing.ZeroHash.String() {
		fmt.Println("Bootstrap not detected. Bootstrapping...")
		err := a.Update()
		if err != nil {
			return nil, err
		}

		fmt.Println("Finished bootstrapping.")
	}
	return a, nil
}

func (a *APM) parseAndRun(
	alias string,
	command func(string) error,
) error {
	if qualifiedName(alias) {
		return command(alias)
	}

	fullName, err := a.getFullNameForAlias(alias)
	if err != nil {
		return err
	}

	return command(fullName)
}

func (a *APM) Install(alias string) error {
	return a.parseAndRun(alias, a.install)
}

func (a *APM) install(name string) error {
	if err := a.lock.TryLock(); err != nil {
		return err
	}
	defer func() {
		_ = a.lock.Unlock()
	}()

	_, ok := a.stateFile.InstallationRegistry[name]
	if ok {
		fmt.Printf("VM %s is already installed. Skipping.\n", name)
		return nil
	}

	repoAlias, plugin := util.ParseQualifiedName(name)
	organization, repo := util.ParseAlias(repoAlias)

	repository, err := a.repoFactory.GetRepository(repoAlias)
	if err != nil {
		return err
	}

	workflow := workflow.NewInstall(workflow.InstallConfig{
		Name:         name,
		Plugin:       plugin,
		Organization: organization,
		Repo:         repo,
		TmpPath:      a.tmpPath,
		PluginPath:   a.pluginPath,
		StateFile:    a.stateFile,
		Repository:   repository,
		Fs:           a.fs,
		Installer:    a.installer,
	})

	return a.executor.Execute(workflow)
}

func (a *APM) Uninstall(alias string) error {
	return a.parseAndRun(alias, a.uninstall)
}

func (a *APM) uninstall(name string) error {
	if err := a.lock.TryLock(); err != nil {
		return err
	}
	defer func() {
		_ = a.lock.Unlock()
	}()

	alias, plugin := util.ParseQualifiedName(name)
	wf := workflow.NewUninstall(
		workflow.UninstallConfig{
			Name:       name,
			Plugin:     plugin,
			RepoAlias:  alias,
			StateFile:  a.stateFile,
			Fs:         a.fs,
			PluginPath: a.pluginPath,
		},
	)

	return a.executor.Execute(wf)
}

func (a *APM) JoinSubnet(alias string) error {
	return a.parseAndRun(alias, a.joinSubnet)
}

func (a *APM) joinSubnet(fullName string) error {
	alias, plugin := util.ParseQualifiedName(fullName)
	repo, err := a.repoFactory.GetRepository(alias)
	if err != nil {
		return err
	}

	definition, err := repo.GetSubnet(plugin)
	if err != nil {
		return err
	}

	subnet := definition.Definition
	subnetID, _ := subnet.GetID(constant.DefaultNetwork)

	// TODO prompt user, add force flag
	fmt.Printf("Installing virtual machines for subnet %s.\n", subnetID)
	for _, vm := range subnet.VMs {
		if err := a.Install(strings.Join([]string{alias, vm}, constant.QualifiedNameDelimiter)); err != nil {
			return err
		}
	}

	fmt.Printf("Updating virtual machines...\n")
	if err := a.adminClient.LoadVMs(); errors.Is(err, syscall.ECONNREFUSED) {
		fmt.Printf("Node at %s was offline. Virtual machines will be available upon node startup.\n", a.adminAPIEndpoint)
	} else if err != nil {
		return err
	}

	fmt.Printf("Whitelisting subnet %s...\n", subnetID)
	if err := a.adminClient.WhitelistSubnet(subnetID); errors.Is(err, syscall.ECONNREFUSED) {
		fmt.Printf("Node at %s was offline. You'll need to whitelist the subnet upon node restart.\n", a.adminAPIEndpoint)
	} else if err != nil {
		return err
	}

	fmt.Printf("Finished installing virtual machines for subnet %s.\n", subnet.ID)
	return nil
}

func (a *APM) Info(alias string) error {
	if qualifiedName(alias) {
		return a.install(alias)
	}

	fullName, err := a.getFullNameForAlias(alias)
	if err != nil {
		return err
	}

	return a.info(fullName)
}

func (a *APM) info(_ string) error {
	return nil
}

func (a *APM) Update() error {
	if err := a.lock.TryLock(); err != nil {
		return err
	}
	defer func() {
		_ = a.lock.Unlock()
	}()

	workflow := workflow.NewUpdate(workflow.UpdateConfig{
		Executor:         a.executor,
		StateFile:        a.stateFile,
		TmpPath:          a.tmpPath,
		PluginPath:       a.pluginPath,
		Installer:        a.installer,
		RepositoriesPath: a.repositoriesPath,
		Auth:             a.auth,
		RepoFactory:      a.repoFactory,
		Fs:               a.fs,
		Git:              a.git,
	})

	if err := a.executor.Execute(workflow); err != nil {
		return err
	}

	return nil
}

func (a *APM) Upgrade(alias string) error {
	if err := a.lock.TryLock(); err != nil {
		return err
	}
	defer func() {
		_ = a.lock.Unlock()
	}()

	// If we have an alias specified, upgrade the specified VM.
	if alias != "" {
		return a.parseAndRun(alias, a.upgradeVM)
	}

	// Otherwise, just upgrade everything.
	wf := workflow.NewUpgrade(workflow.UpgradeConfig{
		Executor:    a.executor,
		RepoFactory: a.repoFactory,
		StateFile:   a.stateFile,
		TmpPath:     a.tmpPath,
		PluginPath:  a.pluginPath,
		Installer:   a.installer,
		Fs:          a.fs,
		Git:         a.git,
	})

	return a.executor.Execute(wf)
}

func (a *APM) upgradeVM(name string) error {
	return a.executor.Execute(workflow.NewUpgradeVM(
		workflow.UpgradeVMConfig{
			Executor:    a.executor,
			FullVMName:  name,
			RepoFactory: a.repoFactory,
			StateFile:   a.stateFile,
			TmpPath:     a.tmpPath,
			PluginPath:  a.pluginPath,
			Installer:   a.installer,
			Fs:          a.fs,
			Git:         a.git,
		},
	))
}

func (a *APM) AddRepository(alias string, url string, branch string) error {
	if err := a.lock.TryLock(); err != nil {
		return err
	}
	defer func() {
		_ = a.lock.Unlock()
	}()

	if !util.ValidAlias(alias) {
		return fmt.Errorf("%s is not a valid alias (must be in the form of organization/repository)", alias)
	}

	wf := workflow.NewAddRepository(
		workflow.AddRepositoryConfig{
			SourcesList: a.stateFile.Sources,
			Alias:       alias,
			URL:         url,
			Branch:      plumbing.NewBranchReferenceName(branch),
		},
	)

	return a.executor.Execute(wf)
}

func (a *APM) RemoveRepository(alias string) error {
	if err := a.lock.TryLock(); err != nil {
		return err
	}
	defer func() {
		_ = a.lock.Unlock()
	}()

	return a.executor.Execute(workflow.NewRemoveRepository(
		workflow.RemoveRepositoryConfig{
			SourcesList:      a.stateFile.Sources,
			RepositoriesPath: a.repositoriesPath,
			Alias:            alias,
		},
	))
}

func (a *APM) ListRepositories() error {
	if err := a.lock.TryLock(); err != nil {
		return err
	}
	defer func() {
		_ = a.lock.Unlock()
	}()

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "alias\turl\tbranch")
	for alias, metadata := range a.stateFile.Sources {
		fmt.Fprintf(w, "%s\t%s\t%s\n", alias, metadata.URL, metadata.Branch)
	}
	w.Flush()
	return nil
}

func qualifiedName(name string) bool {
	parsed := strings.Split(name, ":")
	return len(parsed) > 1
}

func (a *APM) getFullNameForAlias(alias string) (string, error) {
	matches := make([]string, 0)

	for alias := range a.stateFile.Sources {
		// See if this repo exists
		_, err := a.repoFactory.GetRepository(alias)
		if err != nil {
			return "", err
		}

		matches = append(matches, alias)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("more than one match found for %s. Please specify the fully qualified name. Matches: %s", alias, matches)
	}

	return fmt.Sprintf("%s:%s", matches[0], alias), nil
}
