package magicnumberr_test

import (
	"os"
	"testing"

	"github.com/Defacto2/server/internal/magicnumberr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIcon(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(icoFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Ico(r))
	assert.Equal(t, magicnumberr.MicrosoftIcon, magicnumberr.Find(r))
}

func TestAVIF(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(avifFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Avif(r))
	assert.Equal(t, magicnumberr.AV1ImageFile, magicnumberr.Find(r))
}

func TestBMP(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(bmpFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Bmp(r))
	assert.Equal(t, magicnumberr.BMPFileFormat, magicnumberr.Find(r))
}

func TestGif(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(gifFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Gif(r))
	assert.Equal(t, magicnumberr.GraphicsInterchangeFormat, magicnumberr.Find(r))
	r, err = os.Open(uncompress(gif2File))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Gif(r))
	assert.Equal(t, magicnumberr.GraphicsInterchangeFormat, magicnumberr.Find(r))
}

func TestIlbm(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(ilbmFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Ilbm(r))
	assert.Equal(t, magicnumberr.InterleavedBitmap, magicnumberr.Find(r))
}

func TestJpeg(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(jpgFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Jpeg(r))
	assert.Equal(t, magicnumberr.JPEGFileInterchangeFormat, magicnumberr.Find(r))
	r, err = os.Open(uncompress(jpegFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Jpeg(r))
	assert.Equal(t, magicnumberr.JPEGFileInterchangeFormat, magicnumberr.Find(r))
}

func TestPCX(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(pcxFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Pcx(r))
	assert.Equal(t, magicnumberr.PersonalComputereXchange, magicnumberr.Find(r))
}

func TestPNG(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(pngFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Png(r))
	assert.Equal(t, magicnumberr.PortableNetworkGraphics, magicnumberr.Find(r))
	sign, err := magicnumberr.Image(r)
	require.NoError(t, err)
	assert.Equal(t, magicnumberr.PortableNetworkGraphics, sign)

}

func TestWebp(t *testing.T) {
	t.Parallel()
	r, err := os.Open(uncompress(webpFile))
	assert.NoError(t, err)
	defer r.Close()
	assert.True(t, magicnumberr.Webp(r))
	assert.Equal(t, magicnumberr.GoogleWebP, magicnumberr.Find(r))
}
