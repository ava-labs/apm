package storage

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
)

import (
	"gopkg.in/yaml.v3"
)

var (
	checkpointsPrefix                     = []byte("checkpoint")
	_                 Storage[SourceInfo] = &SourceDB{}
)

func NewSourceDB(config SourceDBConfig) *SourceDB {
	sources := &SourceDB{
		db: prefixdb.New(checkpointsPrefix, config.DB),
	}
	return sources
}

type SourceDBConfig struct {
	DB database.Database
}

type SourceDB struct {
	db database.Database
}

func (c *SourceDB) Put(bytes []byte, info SourceInfo) error {
	updatedInfoBytes, err := yaml.Marshal(info)
	if err != nil {
		return err
	}
	return c.db.Put(bytes, updatedInfoBytes)
}

func (c *SourceDB) Get(alias []byte) (SourceInfo, error) {
	sourceInfoBytes, err := c.db.Get(alias)
	if err != nil {
		return SourceInfo{}, err
	}

	sourceInfo := SourceInfo{}
	if err := yaml.Unmarshal(sourceInfoBytes, &sourceInfo); err != nil {
		return SourceInfo{}, err
	}

	return sourceInfo, nil
}

func (c *SourceDB) Delete(bytes []byte) error {
	return c.db.Delete(bytes)
}

func (c *SourceDB) Iterator() database.Iterator {
	return c.db.NewIterator()
}
