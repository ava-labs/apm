package definition

import (
	"github.com/ava-labs/avalanchego/chains"
	"github.com/ava-labs/avalanchego/version"
)

type Subnet interface {
	Plugin

	// VMs that are required to process this subnet
	VMs() map[VM]version.Version
	// SubnetConfig describes parameters for the subnet.
	SubnetConfig() *chains.SubnetConfig
}
