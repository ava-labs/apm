package types

import "github.com/ava-labs/avalanchego/chains"

type Subnet struct {
	ID_            string              `yaml:"id"`
	Alias_         string              `yaml:"alias"`
	Homepage_      string              `yaml:"homepage"`
	Description_   string              `yaml:"description"`
	Maintainers_   []string            `yaml:"maintainers"`
	InstallScript_ string              `yaml:"installScript"`
	VMs_           []string            `yaml:"vms"`
	Config_        chains.SubnetConfig `yaml:"config"`
}

func (s *Subnet) ID() string {
	return s.ID_
}

func (s *Subnet) Alias() string {
	return s.Alias_
}

func (s *Subnet) Homepage() string {
	return s.Homepage_
}

func (s *Subnet) Description() string {
	return s.Description_
}

func (s *Subnet) Maintainers() []string {
	return s.Maintainers_
}

func (s *Subnet) InstallScript() string {
	return s.InstallScript_
}
