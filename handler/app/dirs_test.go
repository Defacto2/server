package app_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/stretchr/testify/require"
)

func TestArtifact404(t *testing.T) {
	t.Parallel()
	err := app.Artifact404(newContext(), "")
	require.Error(t, err)
}

func TestArtifact(t *testing.T) {
	t.Parallel()
	dir := app.Dirs{}
	err := dir.Artifact(newContext(), nil, nil, false)
	require.Error(t, err)
}

func TestEditor(t *testing.T) {
	t.Parallel()
	dir := app.Dirs{}
	x := dir.Editor(nil, nil)
	require.Empty(t, x)
}

func TestFileMissingErr(t *testing.T) {
	t.Parallel()
	err := app.FileMissingErr(newContext(), "", nil)
	require.Error(t, err)
}

func TestForbiddenErr(t *testing.T) {
	t.Parallel()
	err := app.ForbiddenErr(newContext(), "", nil)
	require.Error(t, err)
}

func TestInternalErr(t *testing.T) {
	t.Parallel()
	err := app.InternalErr(newContext(), "", nil)
	require.Error(t, err)
}

func TestStatusErr(t *testing.T) {
	t.Parallel()
	err := app.StatusErr(newContext(), -1, "")
	require.Error(t, err)
}
