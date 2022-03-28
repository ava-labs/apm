package definition

import "github.com/ava-labs/avalanchego/chains"

type Subnet interface {
	Plugin

	// VMs that are required to process this subnet
	VMs() []VM
	// SubnetConfig describes parameters for the subnet.
	SubnetConfig() *chains.SubnetConfig
}
