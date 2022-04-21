package types

import "github.com/ava-labs/avalanchego/version"

var _ Plugin = &VM{}

type VM struct {
	ID_           string                  `yaml:"id"`
	Alias_        string                  `yaml:"alias"`
	Homepage_     string                  `yaml:"homepage"`
	Description_  string                  `yaml:"description"`
	Maintainers_  []string                `yaml:"maintainers"`
	InstallScript string                  `yaml:"installScript"`
	BinaryPath    string                  `yaml:"binaryPath"`
	URL           string                  `yaml:"url"`
	SHA256        string                  `yaml:"sha256"`
	Version       version.SemanticVersion `yaml:"version"`
}

func (vm *VM) ID() string {
	return vm.ID_
}

func (vm *VM) Alias() string {
	return vm.Alias_
}

func (vm *VM) Homepage() string {
	return vm.Homepage_
}

func (vm *VM) Description() string {
	return vm.Description_
}

func (vm *VM) Maintainers() []string {
	return vm.Maintainers_
}
