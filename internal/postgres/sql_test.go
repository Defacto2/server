package postgres_test

import (
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/stretchr/testify/assert"
)

func TestVersion_Query(t *testing.T) {
	var v postgres.Version
	err := v.Query()
	assert.Error(t, err)
	assert.Empty(t, v)
}

func TestVersion_String(t *testing.T) {
	var v postgres.Version
	assert.Empty(t, v.String())
}

func TestRole_Role(t *testing.T) {
	r := postgres.Roles()
	assert.NotEmpty(t, r)
}

func TestRole_Select(t *testing.T) {
	var r postgres.Role
	s := r.Select()
	assert.Contains(t, s, "SELECT DISTINCT")
}

func TestDist(t *testing.T) {
	s := postgres.DistScener()
	assert.Contains(t, s, "scener")

	s = postgres.DistWriter()
	assert.Contains(t, s, "credit_text")
	s = postgres.DistArtist()
	assert.Contains(t, s, "credit_illustration")
	s = postgres.DistCoder()
	assert.Contains(t, s, "credit_program")
	s = postgres.DistMusician()
	assert.Contains(t, s, "credit_audio")
	s = postgres.DistMagazine()
	assert.Contains(t, s, "magazine")

	s = postgres.DistReleaser()
	assert.Contains(t, s, "BBS")
	assert.Contains(t, s, "FTP")

	s = postgres.DistReleaserSummed()
	assert.Contains(t, s, "sub.count_sum")

	s = postgres.DistBBSSummed()
	assert.Contains(t, s, "sub.count_sum")

	s = postgres.SumReleaser("")
	assert.Contains(t, s, "")
	s = postgres.SumReleaser("magazine")
	assert.Contains(t, s, "'magazine'")

}
