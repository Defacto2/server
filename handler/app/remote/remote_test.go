package remote_test

// This is a test file is to confirm there's no panics with nil values.

import (
	"testing"

	"github.com/Defacto2/server/handler/app/remote"
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
