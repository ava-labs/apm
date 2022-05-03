package repository

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
)

var (
	repo   = []byte("repo")
	vm     = []byte("vm")
	subnet = []byte("subnet")

	_ Group = &PluginGroup{}
)

// Group defines access to a plugin repository's plugins
type Group interface {
	VMs() database.Database
	Subnets() database.Database
}

// PluginGroupConfig configures parameters for PluginGroup
type PluginGroupConfig struct {
	Alias []byte
	DB    database.Database
}

// PluginGroup wraps a plugin repository's vms and subnets
type PluginGroup struct {
	vms, subnets database.Database
}

// NewPluginGroup returns an instance of PluginGroup
func NewPluginGroup(config PluginGroupConfig) *PluginGroup {
	// all repositories
	repositories := prefixdb.New(repo, config.DB)
	// this specific repository
	repo := prefixdb.New(config.Alias, repositories)

	// vms and subnets for this repository
	repoVMs := prefixdb.New(vm, repo)
	repoSubnets := prefixdb.New(subnet, repo)

	return &PluginGroup{
		vms:     repoVMs,
		subnets: repoSubnets,
	}
}

func (p *PluginGroup) VMs() database.Database {
	return p.vms
}

func (p *PluginGroup) Subnets() database.Database {
	return p.subnets
}
