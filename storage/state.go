package storage

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanchego/utils/perms"
	"gopkg.in/yaml.v3"
)

const (
	stateFile = "apm.state"
)

func NewEmptyStateFile(path string) StateFile {
	return StateFile{
		Sources:      make(map[string]SourceInfo),
		RepoList:     make(map[string]RepoList),
		InstalledVMs: make(map[string]InstallInfo),
		path:         filepath.Join(path, stateFile),
	}
}

func NewStateFile(path string) (StateFile, error) {
	result := NewEmptyStateFile(path)

	b, err := os.ReadFile(result.path)
	var pathError *os.PathError
	if errors.As(err, &pathError) {
		// need to initialize the StateFile on Commit
	} else if err != nil {
		return StateFile{}, err
	}

	if err := yaml.Unmarshal(b, result); err != nil {
		return StateFile{}, err
	}

	return result, nil
}

// StateFile is the representation of the current APM state.
// Not safe for concurrent use.
type StateFile struct {
	// Mapping of each tracked repository's alias to its metadata
	Sources map[string]SourceInfo `yaml:"sources,omitempty"`
	// Mapping of each registered vm/subnet and which repos it exists in
	RepoList map[string]RepoList `yaml:"repoList,omitempty"`
	// Mapping of each installed vm's alias to the version installed
	InstalledVMs map[string]InstallInfo `yaml:"installedVMs,omitempty"`

	path string
}

func (s *StateFile) Commit() error {
	bytes, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, bytes, perms.ReadWrite)
}
