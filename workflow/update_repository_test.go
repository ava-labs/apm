// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"path/filepath"
	"testing"

	"github.com/ava-labs/avalanchego/utils/perms"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/apm/storage"
	mockdb "github.com/ava-labs/apm/storage/mocks"
	"github.com/ava-labs/apm/types"
)

func TestUpdateRepositoryExecute(t *testing.T) {
	const (
		repoName       = "repoName"
		repositoryPath = "repository"

		alias = "organization/repository"
		url   = "url"

		tmpPath    = "tmp"
		pluginPath = "pluginDir"

		spacesVM     = "spacesvm"
		spacesSubnet = "spaces"
	)
	var (
		subnetsPath = filepath.Join(repositoryPath, "subnets")
		vmsPath     = filepath.Join(repositoryPath, "vms")

		aliasBytes     = []byte(alias)
		previousCommit = plumbing.Hash{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		latestCommit   = plumbing.Hash{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
		sourceInfo     = storage.SourceInfo{
			Alias:  alias,
			URL:    url,
			Commit: previousCommit,
		}
	)

	// don't try to reformat this; yaml is whitespace sensitive.
	vm := []byte(`vm:
  id: "sqja3uK17MJxfC7AN8nGadBw9JK5BcrsNwNynsqP5Gih8M5Bm"
  alias: "spacesvm"
  homepage: "https://tryspaces.xyz"
  description: "Virtual machine that processes the spaces subnet."
  maintainers:
    - "patrickogrady@avalabs.org"
  installScript: "scripts/build.sh"
  binaryPath: "build/sqja3uK17MJxfC7AN8nGadBw9JK5BcrsNwNynsqP5Gih8M5Bm"
  url: "https://github.com/ava-labs/spacesvm/archive/refs/tags/v0.0.3.tar.gz"
  sha256: "1ac250f6c40472f22eaf0616fc8c886078a4eaa9b2b85fbb4fb7783a1db6af3f"
  version:
    major: 0
    minor: 0
    patch: 3`,
	)

	subnet := []byte(`subnet:
  id: "Ai42MkKqk8yjXFCpoHXw7rdTWSHiKEMqh5h8gbxwjgkCUfkrk"
  alias: "spaces"
  homepage: "https://tryspaces.xyz"
  description: |
    Spaces enables authenticated, hierarchical storage of arbitrary keys/values using any EIP-712 compatible wallet.
  maintainers:
    - "patrickogrady@avalabs.org"
  installScript: "" # no install script needed
  vms:
    - "spacesvm"
  config:
    # This the default subnet config as of avalanchego v1.7.10
    # TODO remove this
    gossipAcceptedFrontierValidatorSize: 0
    gossipAcceptedFrontierNonValidatorSize: 0
    gossipAcceptedFrontierPeerSize: 35
    gossipOnAcceptValidatorSize: 0
    gossipOnAcceptNonValidatorSize: 0
    gossipOnAcceptPeerSize: 20
    appGossipValidatorSize: 10
    appGossipNonValidatorSize: 0
    appGossipPeerSize: 0
    validatorOnly: false
    consensusParameters:
      k: 20
      alpha: 15
      betaVirtuous: 15
      betaRogue: 20
      concurrentRepolls: 4
      optimalProcessing: 50
      maxOutstandingItems: 1024
      maxItemProcessingTime: 120_000_000_000
      parents: 5
      batchSize: 30`,
	)

	setupFs := func(fs afero.Fs) {
		errs := wrappers.Errs{}

		errs.Add(
			fs.MkdirAll(tmpPath, perms.ReadWrite),
			fs.MkdirAll(pluginPath, perms.ReadWrite),
			fs.MkdirAll(subnetsPath, perms.ReadWrite),
			fs.MkdirAll(vmsPath, perms.ReadWrite),
		)

		if errs.Errored() {
			t.Fatal(errs.Err)
		}
	}

	type mocks struct {
		ctrl *gomock.Controller

		fs afero.Fs

		registry     *storage.MockStorage[storage.RepoList]
		sourcesList  *storage.MockStorage[storage.SourceInfo]
		installedVMs *storage.MockStorage[storage.InstallInfo]
		vms          *storage.MockStorage[storage.Definition[types.VM]]
		subnets      *storage.MockStorage[storage.Definition[types.Subnet]]

		installer *MockInstaller

		executor *MockExecutor
	}
	tests := []struct {
		name    string
		setup   func(*testing.T, mocks)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success: vm definitions updated",
			setup: func(t *testing.T, mocks mocks) {
				setupFs(mocks.fs)
				assert.Nil(t, afero.WriteFile(mocks.fs, filepath.Join(vmsPath, "vm-1.yaml"), vm, perms.ReadWrite))

				// update subnet definitions
				mocks.registry.EXPECT().Get([]byte(spacesVM)).Return(storage.RepoList{Repositories: []string{}}, nil)
				mocks.registry.EXPECT().Put([]byte(spacesVM), storage.RepoList{Repositories: []string{alias}}).Return(nil)
				mocks.vms.EXPECT().Put([]byte(spacesVM), gomock.Any()).Return(nil) // TODO fix

				mocks.vms.EXPECT().Iterator().DoAndReturn(func() storage.Iterator[storage.Definition[types.VM]] {
					itr := mockdb.NewMockIterator(mocks.ctrl)
					defer itr.EXPECT().Release()
					itr.EXPECT().Next().Return(false)

					return *storage.NewIterator[storage.Definition[types.VM]](itr)
				})
				mocks.subnets.EXPECT().Iterator().DoAndReturn(func() storage.Iterator[storage.Definition[types.Subnet]] {
					itr := mockdb.NewMockIterator(mocks.ctrl)
					defer itr.EXPECT().Release()
					itr.EXPECT().Next().Return(false)

					return *storage.NewIterator[storage.Definition[types.Subnet]](itr)
				})
				mocks.installedVMs.EXPECT().Iterator().DoAndReturn(func() storage.Iterator[storage.InstallInfo] {
					itr := mockdb.NewMockIterator(mocks.ctrl)
					defer itr.EXPECT().Release()
					itr.EXPECT().Next().Return(false)

					return *storage.NewIterator[storage.InstallInfo](itr)
				})
				mocks.sourcesList.EXPECT().Put([]byte(alias), storage.SourceInfo{
					Alias:  alias,
					URL:    url,
					Commit: latestCommit,
				})
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "success: subnet definitions updated",
			setup: func(t *testing.T, mocks mocks) {
				setupFs(mocks.fs)
				assert.Nil(t, afero.WriteFile(mocks.fs, filepath.Join(subnetsPath, "subnet-1.yaml"), subnet, perms.ReadWrite))

				// update subnet definitions
				mocks.registry.EXPECT().Get([]byte(spacesSubnet)).Return(storage.RepoList{Repositories: []string{}}, nil)
				mocks.registry.EXPECT().Put([]byte(spacesSubnet), storage.RepoList{Repositories: []string{alias}}).Return(nil)
				mocks.subnets.EXPECT().Put([]byte(spacesSubnet), gomock.Any()).Return(nil) // TODO fix

				mocks.vms.EXPECT().Iterator().DoAndReturn(func() storage.Iterator[storage.Definition[types.VM]] {
					itr := mockdb.NewMockIterator(mocks.ctrl)
					defer itr.EXPECT().Release()
					itr.EXPECT().Next().Return(false)

					return *storage.NewIterator[storage.Definition[types.VM]](itr)
				})
				mocks.subnets.EXPECT().Iterator().DoAndReturn(func() storage.Iterator[storage.Definition[types.Subnet]] {
					itr := mockdb.NewMockIterator(mocks.ctrl)
					defer itr.EXPECT().Release()
					itr.EXPECT().Next().Return(false)

					return *storage.NewIterator[storage.Definition[types.Subnet]](itr)
				})
				mocks.installedVMs.EXPECT().Iterator().DoAndReturn(func() storage.Iterator[storage.InstallInfo] {
					itr := mockdb.NewMockIterator(mocks.ctrl)
					defer itr.EXPECT().Release()
					itr.EXPECT().Next().Return(false)

					return *storage.NewIterator[storage.InstallInfo](itr)
				})
				mocks.sourcesList.EXPECT().Put([]byte(alias), storage.SourceInfo{
					Alias:  alias,
					URL:    url,
					Commit: latestCommit,
				})
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fs := afero.NewMemMapFs()

			executor := NewMockExecutor(ctrl)
			installer := NewMockInstaller(ctrl)

			var (
				registry     *storage.MockStorage[storage.RepoList]
				sourcesList  *storage.MockStorage[storage.SourceInfo]
				installedVMs *storage.MockStorage[storage.InstallInfo]
				vms          *storage.MockStorage[storage.Definition[types.VM]]
				subnets      *storage.MockStorage[storage.Definition[types.Subnet]]
			)

			registry = storage.NewMockStorage[storage.RepoList](ctrl)
			sourcesList = storage.NewMockStorage[storage.SourceInfo](ctrl)
			installedVMs = storage.NewMockStorage[storage.InstallInfo](ctrl)
			vms = storage.NewMockStorage[storage.Definition[types.VM]](ctrl)
			subnets = storage.NewMockStorage[storage.Definition[types.Subnet]](ctrl)

			repository := storage.Repository{
				VMs:     vms,
				Subnets: subnets,
			}

			test.setup(t, mocks{
				ctrl:         ctrl,
				fs:           fs,
				registry:     registry,
				sourcesList:  sourcesList,
				installedVMs: installedVMs,
				executor:     executor,
				installer:    installer,
				vms:          vms,
				subnets:      subnets,
			})

			wf := NewUpdateRepository(
				UpdateRepositoryConfig{
					Executor:       executor,
					RepoName:       repoName,
					RepositoryPath: repositoryPath,
					AliasBytes:     aliasBytes,
					PreviousCommit: previousCommit,
					LatestCommit:   latestCommit,
					SourceInfo:     sourceInfo,
					Repository:     repository,
					Registry:       registry,
					SourcesList:    sourcesList,
					InstalledVMs:   installedVMs,
					TmpPath:        tmpPath,
					PluginPath:     pluginPath,
					Installer:      installer,
					Fs:             fs,
				},
			)

			test.wantErr(t, wf.Execute())
		})
	}
}
