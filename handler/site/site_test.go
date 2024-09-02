package site_test

import (
	"testing"

	"github.com/Defacto2/server/handler/site"
	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	t.Parallel()
	website := site.Find("defacto2")
	assert.NotEmpty(t, website)
	assert.Len(t, website, 1)
	assert.Equal(t, "https://defacto2.net", website[0].URL)
	assert.Equal(t, "Defacto2", website[0].Name)
	assert.False(t, website[0].NotWorking)

	website = site.Find("notfound")
	assert.Empty(t, website)

	website = site.Find("razor-1911-demo")
	assert.NotEmpty(t, website)
	assert.Len(t, website, 2)
}
