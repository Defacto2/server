package model_test

import (
	"testing"

	"github.com/Defacto2/server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestGetPlatformTagInfo(t *testing.T) {
	t.Parallel()

	s, err := model.GetPlatformTagInfo("", "")
	require.Error(t, err)
	assert.Empty(t, s)

	s, err = model.GetPlatformTagInfo("ansi", "bbs")
	require.NoError(t, err)
	assert.NotEmpty(t, s)
}

func TestGetTagInfo(t *testing.T) {
	t.Parallel()

	s, err := model.GetTagInfo("")
	require.Error(t, err)
	assert.Empty(t, s)

	s, err = model.GetTagInfo("ansi")
	require.NoError(t, err)
	assert.NotEmpty(t, s)
}

func TestUpdateOnline(t *testing.T) {
	t.Parallel()

	err := model.UpdateOnline(-1)
	require.Error(t, err)
}

func TestUpdateOffline(t *testing.T) {
	t.Parallel()

	err := model.UpdateOffline(-1)
	require.Error(t, err)
}

func TestUpdateNoReadme(t *testing.T) {
	t.Parallel()

	err := model.UpdateNoReadme(-1, false)
	require.Error(t, err)
}

func TestUpdatePlatform(t *testing.T) {
	t.Parallel()

	err := model.UpdatePlatform(-1, "")
	require.ErrorIs(t, err, model.ErrPlatform)

	err = model.UpdatePlatform(-1, "ansi")
	require.Error(t, err)
}

func TestUpdateTag(t *testing.T) {
	t.Parallel()

	err := model.UpdateTag(-1, "")
	require.ErrorIs(t, err, model.ErrTag)

	err = model.UpdateTag(-1, "bbs")
	require.Error(t, err)
}

func TestUpdateTitle(t *testing.T) {
	t.Parallel()

	err := model.UpdateTitle(-1, "")
	require.Error(t, err)
}

func TestUpdateYMD(t *testing.T) {
	t.Parallel()

	empty := null.Int16{}
	err := model.UpdateYMD(-1, empty, empty, empty)
	require.Error(t, err)

	y := null.Int16From(1900)
	err = model.UpdateYMD(-1, y, empty, empty)
	require.ErrorIs(t, err, model.ErrYear)

	m := null.Int16From(13)
	err = model.UpdateYMD(-1, empty, m, empty)
	require.ErrorIs(t, err, model.ErrMonth)

	d := null.Int16From(999)
	err = model.UpdateYMD(-1, empty, empty, d)
	require.ErrorIs(t, err, model.ErrDay)
}
