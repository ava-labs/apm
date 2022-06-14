// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package types

import "github.com/ava-labs/avalanchego/subnets"

var _ Definition = &Subnet{}

type Subnet struct {
	ID_          string               `yaml:"id"`
	Alias_       string               `yaml:"alias"`
	Homepage_    string               `yaml:"homepage"`
	Description_ string               `yaml:"description"`
	Maintainers_ []string             `yaml:"maintainers"`
	VMs_         []string             `yaml:"vms"`
	Config_      subnets.SubnetConfig `yaml:"config,omitempty"`
}

func (s Subnet) ID() string {
	return s.ID_
}

func (s Subnet) Alias() string {
	return s.Alias_
}

func (s Subnet) Homepage() string {
	return s.Homepage_
}

func (s Subnet) Description() string {
	return s.Description_
}

func (s Subnet) Maintainers() []string {
	return s.Maintainers_
}
