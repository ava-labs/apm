package storage

import (
	"github.com/ava-labs/apm/types"
)

var (
	repositoryPrefix = []byte("repository")
)

// Repository wraps a plugin repository's VMs and Subnets
type Repository struct {
	VMs     Storage[Definition[types.VM]]
	Subnets Storage[Definition[types.Subnet]]
}
