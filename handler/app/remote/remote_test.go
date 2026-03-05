package remote_test

// This is a test file is to confirm there's no panics with nil values.

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/Defacto2/server/handler/app/remote"
	"github.com/nalgeon/be"
)

func TestDownload(t *testing.T) {
	t.Parallel()
	dl := remote.DemozooLink{} //nolint:exhaustruct
	err := dl.Download(nil, nil, "")
	be.Err(t, err)
}

func TestStat(t *testing.T) {
	t.Parallel()
	dl := remote.DemozooLink{} //nolint:exhaustruct
	err := dl.Stat(nil, nil, "")
	be.Err(t, err)
}

func TestArchiveContent(t *testing.T) {
	t.Parallel()
	dl := remote.DemozooLink{} //nolint:exhaustruct
	err := dl.ArchiveContent(nil, nil, "")
	be.Err(t, err)
}

func TestUpdate(t *testing.T) {
	t.Parallel()
	dl := remote.DemozooLink{} //nolint:exhaustruct
	err := dl.Update(nil, nil)
	be.Err(t, err)
}

func TestFixSceneOrg(t *testing.T) {
	t.Parallel()
	s := "http://files.scene.org/view/demos/groups/trsi/ms-dos/trsiscxt.zip"
	w := remote.FixSceneOrg(s)
	be.Equal(t, "https://files.scene.org/get/demos/groups/trsi/ms-dos/trsiscxt.zip", w)
}

func TestGetExampleCom1(t *testing.T) {
	t.Parallel()
	_, err := remote.GetFile5sec("http://example.com")
	got := err == nil || errors.Is(err, context.DeadlineExceeded)
	be.True(t, got)
}

func TestGetExampleCom2(t *testing.T) {
	t.Parallel()
	_, err := remote.GetFile("http://example.com", *http.DefaultClient)
	be.True(t, (err == nil || errors.Is(err, context.DeadlineExceeded)))
}

func TestGetExampleCom3(t *testing.T) {
	t.Parallel()
	_, err := remote.GetFile("://example.com", *http.DefaultClient)
	be.Err(t, err)
}
