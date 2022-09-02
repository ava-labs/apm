// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

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

func newEmpty(path string) File {
	return File{
		Sources:              make(map[string]*SourceInfo),
		InstallationRegistry: make(map[string]*InstallInfo),
		path:                 filepath.Join(path, stateFile),
	}
}

func New(path string) (File, error) {
	result := newEmpty(path)

	b, err := os.ReadFile(result.path)
	if errors.Is(err, os.ErrNotExist) {
		// The statefile doesn't exist, so we should swallow this error
		// and create it when we call Commit()
	} else if err != nil {
		return File{}, err
	}

	if err := yaml.Unmarshal(b, result); err != nil {
		return File{}, err
	}

	return result, nil
}

// File is the representation of the current APM state.
// Not safe for concurrent use.
type File struct {
	// Mapping of each tracked repository's alias to its metadata
	Sources map[string]*SourceInfo `yaml:"sources"`
	// Mapping of each installed vm's alias to the version installed
	InstallationRegistry map[string]*InstallInfo `yaml:"installation-registry"`

	path string
}

func (s *File) Commit() error {
	bytes, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, bytes, perms.ReadWrite)
}
