//go:build ping

package remote_test

import (
	"flag"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app/remote"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

// These tests are not run by default.
//
// To test the pouet fetch file using production record 1, run:
// go test -v -tags=ping -run TestPouet -args -pid=1
//
// Or for Demozoo, run:
// go test -v -tags=ping -run TestDemozoo -args -pid=1
//

var prodID = flag.String("pid", "", "remote production id from cli")

const unid = "01a06fc0-3e62-72b4-b5f9-1d20e0cabec2"

func TestDemozoo(t *testing.T) {
	if *prodID == "" {
		t.Skip("-pid was not used, no demozoo test will run")
	}

	prodID, err := strconv.Atoi(*prodID)
	be.Err(t, err, nil)
	t.Log("will test using demozoo production id:", prodID)

	tempDir := t.TempDir()
	t.Log("using temp directory for download store:", tempDir)

	download := dir.Directory(tempDir)
	// unid := testutil.UID4
	t.Log("using a standard UUID for record:", unid)

	timeout := 60 * time.Second
	t.Log("attempt at download will timeout after:", timeout)

	got := remote.Demozoo(prodID, unid, download, timeout)
	be.Equal(t, got.ID, prodID)
	be.Equal(t, got.UUID, unid)

	tx := testutil.Tx(t)
	sl := slog.Default()
	c := testutil.NewContext(t, "")
	err = got.Download(t.Context(), sl, c, tx)
	be.Err(t, err, nil)
}

func TestPouet(t *testing.T) {
	if *prodID == "" {
		t.Skip("-pid was not used, no pouet test will run")
	}

	prodID, err := strconv.Atoi(*prodID)
	be.Err(t, err, nil)
	t.Log("will test using pouet production id:", prodID)

	tempDir := t.TempDir()
	t.Log("using temp directory for download store:", tempDir)

	download := dir.Directory(tempDir)
	t.Log("using a standard UUID for record:", unid)

	timeout := 60 * time.Second
	t.Log("attempt at download will timeout after:", timeout)

	got := remote.Pouet(prodID, unid, download, timeout)
	be.Equal(t, got.PouetID, prodID)
	be.Equal(t, got.UUID, unid)

	tx := testutil.Tx(t)
	sl := slog.Default()
	c := testutil.NewContext(t, "")
	err = got.Download(t.Context(), sl, c, tx)
	be.Err(t, err, nil)
}
