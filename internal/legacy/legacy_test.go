package legacy_test

import (
	"testing"

	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/legacy"
	"github.com/nalgeon/be"
)

func TestIsArchiveExt(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"demo.zip", "DEMO.ZIP", "demo.lha", "demo.lzh",
		"demo.arc", "demo.arj", "demo.pak", "demo.lbr",
	} {
		be.True(t, legacy.ArchiveExt(name))
	}

	for _, name := range []string{"", "demo", "demo.rar", "demo.7z", "demo.txt"} {
		be.True(t, !legacy.ArchiveExt(name))
	}
}

func TestIsLegacyMagic(t *testing.T) {
	t.Parallel()

	for _, sig := range []magicnumber.Signature{
		magicnumber.NoGatePAK,
		magicnumber.ARChiveSEA,
		magicnumber.YoshiLHA,
		magicnumber.ArchiveRobertJung,
		magicnumber.LibraryUtility,
		magicnumber.PKWAREZipImplode,
		magicnumber.PKWAREZipReduce,
		magicnumber.PKWAREZipShrink,
	} {
		be.True(t, legacy.Archive(sig.Title()))
	}

	be.True(t, !legacy.Archive(nil))
	be.True(t, !legacy.Archive(123))
	be.True(t, !legacy.Archive(""))
	be.True(t, !legacy.Archive(magicnumber.PKWAREZip.Title()))
	be.True(t, !legacy.Archive(magicnumber.RoshalARchive.Title()))
}
