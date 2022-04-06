package repository

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"

	"github.com/ava-labs/avalanche-plugin/avalanche"
)

var (
	vmPrefix     = []byte("vm")
	subnetPrefix = []byte("subnet")
)

type Repository struct {
	name     string
	vmDB     database.Database
	subnetDB database.Database
}

type Config struct {
	Name string
	DB   database.Database
}

func New(config Config) *Repository {
	repo := &Repository{
		name:     config.Name,
		vmDB:     prefixdb.New(vmPrefix, config.DB),
		subnetDB: prefixdb.New(subnetPrefix, config.DB),
	}

	return repo
}

func GetVMs(start string) []avalanche.VM {
	return nil
}

func GetSubnets(start string) []avalanche.Subnet {
	return nil
}
