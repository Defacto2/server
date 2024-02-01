package magic_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/server/internal/magic"
	"github.com/stretchr/testify/assert"
)

func td(name string) string {
	_, b, _, _ := runtime.Caller(0)
	d := filepath.Join(filepath.Dir(b), "../..")
	x := filepath.Join(d, "assets", "testdata", name)
	return x
}
func tduncompress(name string) string {
	_, b, _, _ := runtime.Caller(0)
	d := filepath.Join(filepath.Dir(b), "../..")
	x := filepath.Join(d, "assets", "testdata", "uncompress", name)
	return x
}

func TestANSIMatch(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	assert.NoError(t, err)
	assert.False(t, magic.ANSIMatcher(b))
	b, err = os.ReadFile(tduncompress("TEST.ANS"))
	assert.NoError(t, err)
	assert.True(t, magic.ANSIMatcher(b))
}

func TestArcSeaMatcher(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	assert.NoError(t, err)
	assert.False(t, magic.ArcSeaMatcher(b))

	match := []byte{0x1a, 0x10, 0x00, 0x00, 0x00, 0x00}
	assert.NoError(t, err)
	assert.True(t, magic.ArcSeaMatcher(match))

	b, err = os.ReadFile(td("ARJ310.ARJ"))
	assert.NoError(t, err)
	assert.True(t, magic.ARJMatcher(b))

	match = []byte{0xe9, 0xeb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	assert.True(t, magic.DOSComMatcher(match))
}

func TestInterchangeMatcher(t *testing.T) {
	t.Parallel()
	// TODO create a IFF test file for testing.
}

func TestPCXMatcher(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	assert.NoError(t, err)
	assert.False(t, magic.PCXMatcher(b))

	b, err = os.ReadFile(tduncompress("TEST.PCX"))
	assert.NoError(t, err)
	assert.True(t, magic.PCXMatcher(b))
}
