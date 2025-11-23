package app_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/nalgeon/be"
)

func TestArtifact404(t *testing.T) {
	t.Parallel()
	err := app.Artifact404(newContext(), nil, "")
	be.Err(t, err)
}

func TestArtifact(t *testing.T) {
	t.Parallel()
	dir := app.Dirs{}
	err := dir.Artifact(newContext(), nil, nil, false)
	be.Err(t, err)
}

func TestEditor(t *testing.T) {
	t.Parallel()
	dir := app.Dirs{}
	x := dir.EditorContent(newContext(), nil, -1, nil, nil)
	be.True(t, len(x) == 0)
}

func TestFileMissingErr(t *testing.T) {
	t.Parallel()
	err := app.FileMissingErr(newContext(), nil, "", nil)
	be.Err(t, err)
}

func TestForbiddenErr(t *testing.T) {
	t.Parallel()
	err := app.ForbiddenErr(newContext(), nil, "", nil)
	be.Err(t, err)
}

func TestInternalErr(t *testing.T) {
	t.Parallel()
	err := app.InternalErr(newContext(), nil, "", nil)
	be.Err(t, err)
}

func TestStatusErr(t *testing.T) {
	t.Parallel()
	err := app.StatusErr(newContext(), nil, -1, "")
	be.Err(t, err)
}
