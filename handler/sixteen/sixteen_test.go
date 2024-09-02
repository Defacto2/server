package sixteen_test

import (
	"testing"

	"github.com/Defacto2/server/handler/sixteen"
	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	t.Parallel()
	tag := sixteen.Find("defacto2")
	assert.Equal(t, tag, sixteen.GroupTag("group/defacto 2"))
	tag = sixteen.Find("notfound")
	assert.Equal(t, tag, sixteen.GroupTag(""))
}
