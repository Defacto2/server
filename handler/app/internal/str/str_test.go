package str_test

import (
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app/internal/str"
	"github.com/stretchr/testify/assert"
)

func TestLinkPreviewTip(t *testing.T) {
	t.Parallel()
	s := str.LinkPreviewTip("", "")
	assert.Empty(t, s)
	s = str.LinkPreviewTip(".zip", "windows")
	assert.Empty(t, s)
	s = str.LinkPreviewTip(".txt", "windows")
	assert.Equal(t, "Read this as text", s)
}

func TestUpdated(t *testing.T) {
	t.Parallel()
	s := str.Updated(nil, "")
	assert.Empty(t, s)
	s = str.Updated(time.Now(), "")
	assert.Contains(t, s, "Time less than 5 seconds ago")
}
