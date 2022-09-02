// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/git"
	"github.com/ava-labs/apm/types"
)

var (
	vmDir     = "vms"
	subnetDir = "subnets"

	vmKey     = "vm"
	subnetKey = "subnet"

	extension = "yaml"
)

// Repository wraps a plugin repository's VMs and Subnets
type Repository interface {
	GetPath() string
	GetVM(name string) (Definition[types.VM], error)
	GetSubnet(name string) (Definition[types.Subnet], error)
}

type DiskRepository struct {
	Git  git.Factory
	Path string
}

func (d DiskRepository) GetVM(name string) (Definition[types.VM], error) {
	return get[types.VM](d, vmDir, name, vmKey)
}

func (d DiskRepository) GetSubnet(name string) (Definition[types.Subnet], error) {
	return get[types.Subnet](d, subnetDir, name, subnetKey)
}

func (d DiskRepository) GetPath() string {
	return d.Path
}

func get[T types.Definition](d DiskRepository, dir string, file string, key string) (Definition[T], error) {
	relativePathWithExtension := filepath.Join(dir, fmt.Sprintf("%s.%s", file, extension))
	absolutePathWithExtension := filepath.Join(d.Path, relativePathWithExtension)
	bytes, err := os.ReadFile(absolutePathWithExtension)
	if err != nil {
		return Definition[T]{}, err
	}

	data := make(map[string]T)
	if err := yaml.Unmarshal(bytes, data); err != nil {
		return Definition[T]{}, err
	}

	definition := data[key]
	commit, err := d.Git.GetLastModified(d.Path, relativePathWithExtension)
	if err != nil {
		return Definition[T]{}, err
	}

	return Definition[T]{
		Definition: definition,
		Commit:     commit,
	}, nil
}
