package remote_test

// This is a test file is to confirm there's no panics with nil values.

import (
	"net/http"
	"testing"

	"github.com/Defacto2/server/handler/app/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownload(t *testing.T) {
	dl := remote.DemozooLink{}
	err := dl.Download(nil, nil, "")
	require.Error(t, err)
}

func TestStat(t *testing.T) {
	dl := remote.DemozooLink{}
	err := dl.Stat(nil, nil, "")
	require.Error(t, err)
}

func TestArchiveContent(t *testing.T) {
	dl := remote.DemozooLink{}
	err := dl.ArchiveContent(nil, nil, "")
	require.NoError(t, err)
}

func TestUpdate(t *testing.T) {
	dl := remote.DemozooLink{}
	err := dl.Update(nil, nil)
	require.Error(t, err)
}

func TestFixSceneOrg(t *testing.T) {
	s := "http://files.scene.org/view/demos/groups/trsi/ms-dos/trsiscxt.zip"
	w := remote.FixSceneOrg(s)
	assert.Equal(t, "https://files.scene.org/get/demos/groups/trsi/ms-dos/trsiscxt.zip", w)
}

func TestGetExampleCom(t *testing.T) {
	t.Parallel()
	r, err := remote.GetFile5sec("http://example.com")
	assert.NotEqual(t, "", r.Path)
	assert.Equal(t, "text/html", r.ContentType)
	require.NoError(t, err)
	_, err = remote.GetFile("http://example.com", *http.DefaultClient)
	require.NoError(t, err)
	_, err = remote.GetFile("://example.com", *http.DefaultClient)
	require.Error(t, err)
}
