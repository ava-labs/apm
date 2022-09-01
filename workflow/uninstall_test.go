// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/apm/state"
	"github.com/ava-labs/apm/types"
)

func TestUninstallExecute(t *testing.T) {
	name := "organization/repository:vm"

	definition := state.Definition[types.VM]{
		Definition: types.VM{
			ID:            "id",
			Alias:         "vm",
			Homepage:      "homepage",
			Description:   "description",
			Maintainers:   []string{"joshua", "kim"},
			InstallScript: "./installScript",
			BinaryPath:    "./build/binaryPath",
			URL:           "url",
			SHA256:        "sha256",
		},
		Commit: "commit",
	}

	type mocks struct {
		stateFile state.File
	}
	tests := []struct {
		name    string
		setup   func(mocks)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "vm already uninstalled",
			setup: func(mocks mocks) {
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
		{
			name: "success",
			setup: func(mocks mocks) {
				mocks.stateFile.InstallationRegistry[name] = &state.InstallInfo{
					ID:     definition.Definition.GetID(),
					Commit: definition.Commit,
				}
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stateFile, err := state.New("stateFilePath")
			require.NoError(t, err)

			test.setup(mocks{
				stateFile: stateFile,
			})

			wf := NewUninstall(
				UninstallConfig{
					Name:      "organization/repository:vm",
					Plugin:    "vm",
					RepoAlias: "organization/repository",
					StateFile: stateFile,
					Fs:        afero.NewMemMapFs(),
				},
			)

			test.wantErr(t, wf.Execute())
		})
	}
}
