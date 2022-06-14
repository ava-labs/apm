// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package types

import "github.com/ava-labs/avalanchego/version"

var _ Definition = &VM{}

type VM struct {
	ID            string           `yaml:"id"`
	Alias         string           `yaml:"alias"`
	Homepage      string           `yaml:"homepage"`
	Description   string           `yaml:"description"`
	Maintainers   []string         `yaml:"maintainers"`
	InstallScript string           `yaml:"installScript"`
	BinaryPath    string           `yaml:"binaryPath"`
	URL           string           `yaml:"url"`
	SHA256        string           `yaml:"sha256"`
	Version       version.Semantic `yaml:"version"`
}

func (vm VM) GetID() string {
	return vm.ID
}

func (vm VM) GetAlias() string {
	return vm.Alias
}

func (vm VM) GetHomepage() string {
	return vm.Homepage
}

func (vm VM) GetDescription() string {
	return vm.Description
}

func (vm VM) GetMaintainers() []string {
	return vm.Maintainers
}
