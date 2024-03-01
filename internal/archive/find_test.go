package archive_test

import (
	"testing"

	"github.com/Defacto2/server/internal/archive"
	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadme(t *testing.T) {
	s := archive.Readme("")
	assert.Empty(t, s)

	dir := td("uncompress")
	files, err := helper.Files(dir)
	require.NoError(t, err)
	assert.Len(t, 16, len(files))

	s = archive.Readme("", files...)
	assert.Equal(t, "TEST.NFO", s)

	s = archive.Readme("TEST.ZIP", files...)
	assert.Equal(t, "TEST.NFO", s)
}
