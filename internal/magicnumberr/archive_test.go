package magicnumberr_test

import (
	"os"
	"testing"

	"github.com/Defacto2/server/internal/magicnumberr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	zipReduceFile  = "PKZ90B4.ZIP"
	zipShrinkFile  = "PKZ80A1.ZIP"
	zipImplodeFile = "PKZ110EI.ZIP"
	zipStoreFile   = "PKZ204E0.ZIP"
	freeArcFile    = "TESTfree.arc"
	seaFile        = "ARC521P.ARC"
	arjFile        = "ARJ310.ARJ"
	tarFile        = "TAR135.TAR"
	rarv5File      = "RAR624.RAR"
	gzFile         = "TAR135.GZ"
	b2zFile        = "TEST.tar.bz2"
	lhaFile        = "LHA114.LZH"
	x7zFile        = "TEST.7z"
	xzFile         = "TEST.tar.xz"
)

func TestArchive(t *testing.T) {
	t.Parallel()
	t.Log("TestArchive")
	r, err := os.Open(td(seaFile))
	require.NoError(t, err)
	defer r.Close()
	sign, err := magicnumberr.Archive(r)
	require.NoError(t, err)
	assert.Equal(t, magicnumberr.ARChiveSEA, sign)
	assert.Equal(t, "ARC by SEA", sign.String())
	assert.Equal(t, "Archive by SEA", sign.Title())
}

func TestZipReduce(t *testing.T) {
	t.Parallel()
	t.Log("TestZipReduce")
	r, err := os.Open(td(zipReduceFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.PkReduce(r))
}

func TestZipShrink(t *testing.T) {
	t.Parallel()
	t.Log("TestZipShrink")
	r, err := os.Open(td(zipShrinkFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.PkShrink(r))
}

func TestZipImplode(t *testing.T) {
	t.Parallel()
	t.Log("TestZipImplode")
	r, err := os.Open(td(zipImplodeFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.PkImplode(r))
}

func TestZipStore(t *testing.T) {
	t.Parallel()
	t.Log("TestZipStore")
	r, err := os.Open(td(zipStoreFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Pkzip(r))
}

func TestTar(t *testing.T) {
	t.Parallel()
	t.Log("TestTar")
	r, err := os.Open(td(tarFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Tar(r))
}

func TestRarv5(t *testing.T) {
	t.Parallel()
	t.Log("TestRarv5")
	r, err := os.Open(td(rarv5File))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Rarv5(r))
}

func TestGzip(t *testing.T) {
	t.Parallel()
	t.Log("TestGzip")
	r, err := os.Open(td(gzFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Gzip(r))
	r, err = os.Open(td(b2zFile))
	require.NoError(t, err)
	defer r.Close()
	assert.False(t, magicnumberr.Gzip(r))
}

func TestBzip2(t *testing.T) {
	t.Parallel()
	t.Log("TestBzip2")
	r, err := os.Open(td(b2zFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Bzip2(r))
}

func TestX7z(t *testing.T) {
	t.Parallel()
	t.Log("TestX7z")
	r, err := os.Open(td(x7zFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.X7z(r))
}

func TestXZ(t *testing.T) {
	t.Parallel()
	t.Log("TestXZ")
	r, err := os.Open(td(xzFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.XZ(r))
}

func TestArcFree(t *testing.T) {
	t.Parallel()
	t.Log("TestArcFree")
	r, err := os.Open(td(freeArcFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.ArcFree(r))
	assert.False(t, magicnumberr.ArcSEA(r))
	b, sign, err := magicnumberr.MatchExt(freeArcFile, r)
	require.NoError(t, err)
	assert.True(t, b)
	assert.Equal(t, magicnumberr.FreeArc, sign)
}

func TestArcSEA(t *testing.T) {
	t.Parallel()
	t.Log("TestArcSEA")
	r, err := os.Open(td(seaFile))
	require.NoError(t, err)
	defer r.Close()
	assert.False(t, magicnumberr.ArcFree(r))
	assert.True(t, magicnumberr.ArcSEA(r))
	b, sign, err := magicnumberr.MatchExt(seaFile, r)
	require.NoError(t, err)
	assert.True(t, b)
	assert.Equal(t, magicnumberr.ARChiveSEA, sign)
}

func TestLHA(t *testing.T) {
	t.Parallel()
	t.Log("TestLzhLha")
	r, err := os.Open(td(lhaFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.LzhLha(r))
	sign := magicnumberr.Find(r)
	assert.Equal(t, magicnumberr.YoshiLHA, sign)
	assert.Equal(t, "LHA by Yoshi", sign.String())
	assert.Equal(t, "Yoshi LHA", sign.Title())
	b, sign, err := magicnumberr.MatchExt(lhaFile, r)
	require.NoError(t, err)
	assert.True(t, b)
	assert.Equal(t, magicnumberr.YoshiLHA, sign)
}

func TestArj(t *testing.T) {
	t.Parallel()
	t.Log("TestArj")
	r, err := os.Open(td(arjFile))
	require.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Arj(r))
}
