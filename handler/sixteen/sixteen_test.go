package sixteen_test

import (
	"testing"

	"github.com/Defacto2/server/handler/sixteen"
	"github.com/nalgeon/be"
)

func TestFind(t *testing.T) {
	t.Parallel()
	tag := sixteen.Find("defacto2")
	be.Equal(t, tag, sixteen.GroupTag("group/defacto 2"))
	tag = sixteen.Find("notfound")
	be.Equal(t, tag, sixteen.GroupTag(""))
}
