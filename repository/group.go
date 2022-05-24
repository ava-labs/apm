package repository

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
)

var (
	repo   = []byte("repo")
	vm     = []byte("vm")
	subnet = []byte("subnet")

	_ Registry = &registry{}
)

// Registry defines access to a plugin repository's plugins
type Registry interface {
	VMs() database.Database
	Subnets() database.Database
}

// RegistryConfig configures parameters for registry
type RegistryConfig struct {
	Alias []byte
	DB    database.Database
}

// registry wraps a plugin repository's vms and subnets
type registry struct {
	vms, subnets database.Database
}

// NewRegistry returns an instance of registry
func NewRegistry(config RegistryConfig) *registry {
	// all repositories
	repositories := prefixdb.New(repo, config.DB)
	// this specific repository
	repo := prefixdb.New(config.Alias, repositories)

	// vms and subnets for this repository
	repoVMs := prefixdb.New(vm, repo)
	repoSubnets := prefixdb.New(subnet, repo)

	return &registry{
		vms:     repoVMs,
		subnets: repoSubnets,
	}
}

func (p *registry) VMs() database.Database {
	return p.vms
}

func (p *registry) Subnets() database.Database {
	return p.subnets
}
