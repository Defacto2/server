package remote_test

// This is a test file is to confirm there's no panics with nil values.

import (
	"testing"
	"time"

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
	const testDefaultTimeout = 0
	r, err := remote.GetFile("http://example.com", testDefaultTimeout)
	assert.NotEqual(t, "", r.Path)
	assert.Equal(t, "text/html; charset=UTF-8", r.ContentType)
	require.NoError(t, err)

	const invalidTimeout = 1 * time.Microsecond
	_, err = remote.GetFile("http://example.com", invalidTimeout)
	require.Error(t, err)
}
