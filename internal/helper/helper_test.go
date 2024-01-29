package helper_test

import (
	"embed"
	"testing"

	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata
var testdataFS embed.FS

func TestGetLocalIPs(t *testing.T) {
	t.Parallel()
	ips, err := helper.GetLocalIPs()
	assert.NoError(t, err)
	assert.NotEmpty(t, ips)
	// we can't test the actual IP addresses as they will be different on each machine.
}

func TestGetLocalHosts(t *testing.T) {
	t.Parallel()
	hosts, err := helper.GetLocalHosts()
	assert.NoError(t, err)
	assert.NotEmpty(t, hosts)
	// we can't test the actual host names as they will be different on each machine.
}

func TestIntegrity(t *testing.T) {
	t.Parallel()
	x, err := helper.Integrity("", embed.FS{})
	assert.Error(t, err)
	assert.Empty(t, x)
	x, err = helper.Integrity("nosuchfile", testdataFS)
	assert.Error(t, err)
	assert.Empty(t, x)
	x, err = helper.Integrity("testdata/TEST.DOC", testdataFS)
	assert.NoError(t, err)
	assert.Equal(t, "sha384-5X6isqmILTavQSao9DigKt3O8fX1Hd6hrGJ7pUROFPYWmkKRnFuWwTnjO3h9QkWP", x)
}

func TestIntegrityFile(t *testing.T) {
	t.Parallel()
	x, err := helper.IntegrityFile("")
	assert.Error(t, err)
	assert.Empty(t, x)
	x, err = helper.IntegrityFile("nosuchfile")
	assert.Error(t, err)
	assert.Empty(t, x)
	x, err = helper.IntegrityFile("testdata/TEST.DOC")
	assert.NoError(t, err)
	assert.Equal(t, "sha384-5X6isqmILTavQSao9DigKt3O8fX1Hd6hrGJ7pUROFPYWmkKRnFuWwTnjO3h9QkWP", x)
}

func TestIntegrityBytes(t *testing.T) {
	t.Parallel()
	x := helper.IntegrityBytes(nil)
	assert.Equal(t, "sha384-OLBgp1GsljhM2TJ+sbHjaiH9txEUvgdDTAzHv2P24donTt6/529l+9Ua0vFImLlb", x)
	x = helper.IntegrityBytes([]byte("hello world"))
	assert.Equal(t, "sha384-/b2OdaZ/KfcBpOBAOF4uI5hjA+oQI5IRr5B/y7g1eLPkF8txzmRu/QgZ3YwIjeG9", x)
}
