package magicnumberr_test

import (
	"os"
	"testing"

	"github.com/Defacto2/server/internal/magicnumberr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMod(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(modFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Mod(r))
	assert.Equal(t, magicnumberr.MusicModule, magicnumberr.Find(r))
	assert.False(t, magicnumberr.MTM(r))
	assert.Equal(t, "ProTracker 8-channel song", magicnumberr.MusicTracker(r))
}

func TestXM(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(xmFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.XM(r))
	assert.Equal(t, magicnumberr.MusicModule, magicnumberr.Find(r))
	assert.False(t, magicnumberr.IT(r))
	assert.Equal(t, "extended module tracked music", magicnumberr.MusicTracker(r))
}
