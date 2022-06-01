package storage

import (
	"testing"

	"github.com/ava-labs/avalanchego/database/memdb"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/apm/types"
)

func TestRepository(t *testing.T) {
	t.Skip("Skipping test...") // TODO remove

	db := memdb.NewWithSize(5)

	// repositoryDB := prefixdb.New([]byte("repository"), db)
	// foobarDB := prefixdb.New([]byte("foobar"), repositoryDB)
	// expectedVMDB := NewVM(foobarDB)
	// expectedSubnetDB := NewSubnet(foobarDB)

	repository := NewRepository(RepositoryConfig{
		Alias: []byte("foobar"),
		DB:    db,
	})

	k1 := []byte("vmfoobarrepositoryk1")

	k2 := []byte("repositoryfoobarsubnetk2")

	repository.VMs().Put(k1, Definition[types.VM]{})
	repository.Subnets().Put(k2, Definition[types.Subnet]{})

	ok, err := db.Has(k1)
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = db.Has(k2)
	assert.True(t, ok)
	assert.NoError(t, err)
}
