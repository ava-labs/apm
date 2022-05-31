package storage

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"

	"github.com/ava-labs/apm/types"
)

var (
	repositoryPrefix = []byte("repository")

	_ Repository = &repository{}
)

// Repository defines access to a plugin repository's plugins
type Repository interface {
	VMs() Storage[Definition[types.VM]]
	Subnets() Storage[Definition[types.Subnet]]
}

// RepositoryConfig configures parameters for repository
type RepositoryConfig struct {
	Alias []byte
	DB    database.Database
}

// repository wraps a plugin repository's vms and subnets
type repository struct {
	vms     Storage[Definition[types.VM]]
	subnets Storage[Definition[types.Subnet]]
}

// NewRepository returns an instance of repository
func NewRepository(config RepositoryConfig) Repository {
	// all repositories
	reposDB := prefixdb.New(repositoryPrefix, config.DB)
	// this specific repository
	repoDB := prefixdb.New(config.Alias, reposDB)

	return &repository{
		vms:     NewVM(repoDB),
		subnets: NewSubnet(repoDB),
	}
}

func (p *repository) VMs() Storage[Definition[types.VM]] {
	return p.vms
}

func (p *repository) Subnets() Storage[Definition[types.Subnet]] {
	return p.subnets
}
