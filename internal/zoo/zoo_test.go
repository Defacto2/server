package zoo_test

import (
	"testing"

	"github.com/Defacto2/server/internal/zoo"
	"github.com/stretchr/testify/assert"
)

// Set to true to test against the remote servers.
const testRemoteServers = false

func TestDemozoo_Get(t *testing.T) {
	t.Parallel()
	prod := zoo.Demozoo{}
	err := prod.Get(-1)
	assert.Error(t, err)
	assert.ErrorAs(t, err, &zoo.ErrID)

	if !testRemoteServers {
		return
	}

	err = prod.Get(1)
	assert.NoError(t, err)
	assert.ErrorAs(t, err, &zoo.ErrSuccess)
}

func TestFind(t *testing.T) {
	t.Parallel()
	prod := zoo.Find("defacto2")
	want := zoo.GroupID(10000)
	assert.Equal(t, want, prod)

	prod = zoo.Find("notfound")
	assert.Equal(t, prod, zoo.GroupID(0))
}
