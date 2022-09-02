// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package types

var _ Definition = &Subnet{}

type Subnet struct {
	ID          map[string]string `yaml:"id"`
	Alias       string            `yaml:"alias"`
	Homepage    string            `yaml:"homepage"`
	Description string            `yaml:"description"`
	Maintainers []string          `yaml:"maintainers"`
	VMs         []string          `yaml:"vms"`
	// Config      subnets.SubnetConfig `yaml:"config,omitempty"`
}

func (s Subnet) GetID(network string) (string, bool) {
	id, ok := s.ID[network]
	return id, ok
}

func (s Subnet) GetAlias() string {
	return s.Alias
}

func (s Subnet) GetHomepage() string {
	return s.Homepage
}

func (s Subnet) GetDescription() string {
	return s.Description
}

func (s Subnet) GetMaintainers() []string {
	return s.Maintainers
}
