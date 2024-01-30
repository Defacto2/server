package web_test

import (
	"testing"

	"github.com/Defacto2/server/internal/web"
	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	t.Parallel()
	website := web.Find("defacto2")
	assert.NotEmpty(t, website)
	assert.Len(t, website, 1)
	assert.Equal(t, "https://defacto2.net", website[0].URL)
	assert.Equal(t, "Defacto2", website[0].Name)
	assert.False(t, website[0].NotWorking)

	website = web.Find("notfound")
	assert.Empty(t, website)
	assert.Len(t, website, 0)

	website = web.Find("razor-1911-demo")
	assert.NotEmpty(t, website)
	assert.Len(t, website, 2)
}
