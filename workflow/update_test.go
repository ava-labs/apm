package workflow

import (
	"fmt"
	"testing"

	"github.com/ava-labs/avalanchego/version"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/apm/storage"
	mockdb "github.com/ava-labs/apm/storage/mocks"
)

func TestUpdateExecute(t *testing.T) {
	errWrong := fmt.Errorf("something went wrong")

	type mocks struct {
		executor     *MockExecutor
		registry     *storage.MockStorage[storage.RepoList]
		installedVMs *storage.MockStorage[version.Semantic]
		sourcesList  *storage.MockStorage[storage.SourceInfo]
		db           *mockdb.MockDatabase
		installer    *MockInstaller
	}
	tests := []struct {
		name    string
		setup   func(mocks)
		wantErr assert.ErrorAssertionFunc
	}{
		{},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			var registry *storage.MockStorage[storage.RepoList]
			var installedVMs *storage.MockStorage[version.Semantic]
			var sourcesList *storage.MockStorage[storage.SourceInfo]

			executor := NewMockExecutor(ctrl)
			db := mockdb.NewMockDatabase(ctrl)
			installer := NewMockInstaller(ctrl)

			registry = storage.NewMockStorage[storage.RepoList](ctrl)
			installedVMs = storage.NewMockStorage[version.Semantic](ctrl)
			sourcesList = storage.NewMockStorage[storage.SourceInfo](ctrl)

			test.setup(mocks{
				executor:     executor,
				registry:     registry,
				installedVMs: installedVMs,
				sourcesList:  sourcesList,
				db:           db,
				installer:    installer,
			})

			wf := NewUpdate(
				UpdateConfig{
					Executor:         executor,
					Registry:         registry,
					InstalledVMs:     installedVMs,
					SourcesList:      sourcesList,
					DB:               db,
					TmpPath:          "tmpPath",
					PluginPath:       "pluginPath",
					Installer:        installer,
					RepositoriesPath: "",
					Auth:             http.BasicAuth{},
				},
			)
			test.wantErr(t, wf.Execute())
		})
	}
}
