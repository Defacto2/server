// Package postgres_test contains tests for the postgres package.
package postgres_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func logr() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

// TestConnection_Open tests the connection to the database.
// It requires a running PostgreSQL server.
func TestConnection_Open(t *testing.T) {
	c := postgres.Connection{}
	conn, err := c.Open()
	require.NoError(t, err)
	defer conn.Close()
	assert.NotNil(t, conn)
}

func TestConnection_Check(t *testing.T) {
	c := postgres.Connection{}
	err := c.Check(nil, false)
	require.Error(t, err)

	err = c.Check(logr(), false)
	require.NoError(t, err)

	c = postgres.Connection{}
	c.Username = "abcde"
	c.Password = ""
	err = c.Check(logr(), false)
	require.NoError(t, err)

	c = postgres.Connection{}
	c.Username = ""
	c.Password = "password"
	c.NoSSLMode = true
	err = c.Check(logr(), false)
	require.NoError(t, err)
}

func TestConnection_New(t *testing.T) {
	c, err := postgres.New()
	require.NoError(t, err)
	assert.NotNil(t, c)
	s, err := c.Open()
	require.NoError(t, err)
	assert.NotNil(t, s)
	defer s.Close()
}

func Test_ConnectDB(t *testing.T) {
	conn, err := postgres.ConnectDB()
	require.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()
}

func TestConnection_URL(t *testing.T) {
	c := postgres.Connection{}
	c.Protocol = "dbproto"
	c.Username = "dbuser"
	c.Password = "xyz"
	c.HostName = "myserver"
	c.HostPort = 5678
	c.Database = "my_db"
	c.NoSSLMode = true
	assert.Equal(t, "dbproto://dbuser:xyz@myserver:5678/my_db?sslmode=disable", c.URL())
}

func TestConnection_Configuration(t *testing.T) {
	c := postgres.Connection{}
	b := strings.Builder{}
	c.Configurations(&b)
	assert.Contains(t, b.String(),
		"PostgreSQL database connection configuration.")
	b = strings.Builder{}

	// test password masking
	c.Password = "abcdef"
	c.Configurations(&b)
	assert.Contains(t, b.String(),
		"******  Password for the database username.")
}
