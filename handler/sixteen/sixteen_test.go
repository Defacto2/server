package sixteen_test

import (
	"testing"

	"github.com/Defacto2/server/handler/sixteen"
	"github.com/nalgeon/be"
)

func TestFind(t *testing.T) {
	t.Parallel()

	tag := sixteen.Find("defacto2")
	be.Equal(t, tag, sixteen.GroupTag("group/defacto%202"))

	tag = sixteen.Find("razor-1911")
	be.Equal(t, tag, sixteen.GroupTag("tags/content/razor%201911"))

	tag = sixteen.Find("notfound")
	be.Equal(t, tag, sixteen.GroupTag(""))
}
