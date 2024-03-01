package model_test

import (
	"testing"

	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestGetPlatformTagInfo(t *testing.T) {
	t.Parallel()

	c := echo.New().NewContext(nil, nil)
	s, err := model.GetPlatformTagInfo(c, "", "")
	require.Error(t, err)
	assert.Empty(t, s)

	s, err = model.GetPlatformTagInfo(c, "ansi", "bbs")
	require.NoError(t, err)
	assert.NotEmpty(t, s)
}

func TestGetTagInfo(t *testing.T) {
	t.Parallel()

	c := echo.New().NewContext(nil, nil)
	s, err := model.GetTagInfo(c, "")
	require.Error(t, err)
	assert.Empty(t, s)

	s, err = model.GetTagInfo(c, "ansi")
	require.NoError(t, err)
	assert.NotEmpty(t, s)
}

func TestUpdateOnline(t *testing.T) {
	t.Parallel()

	c := echo.New().NewContext(nil, nil)
	err := model.UpdateOnline(c, -1)
	require.Error(t, err)
}

func TestUpdateOffline(t *testing.T) {
	t.Parallel()

	c := echo.New().NewContext(nil, nil)
	err := model.UpdateOffline(c, -1)
	require.Error(t, err)
}

func TestUpdateNoReadme(t *testing.T) {
	t.Parallel()

	c := echo.New().NewContext(nil, nil)
	err := model.UpdateNoReadme(c, -1, false)
	require.Error(t, err)
}

func TestUpdatePlatform(t *testing.T) {
	t.Parallel()

	c := echo.New().NewContext(nil, nil)
	err := model.UpdatePlatform(c, -1, "")
	require.ErrorIs(t, err, model.ErrPlatform)

	err = model.UpdatePlatform(c, -1, "ansi")
	require.Error(t, err)
}

func TestUpdateTag(t *testing.T) {
	t.Parallel()

	c := echo.New().NewContext(nil, nil)
	err := model.UpdateTag(c, -1, "")
	require.ErrorIs(t, err, model.ErrTag)

	err = model.UpdateTag(c, -1, "bbs")
	require.Error(t, err)
}

func TestUpdateTitle(t *testing.T) {
	t.Parallel()

	c := echo.New().NewContext(nil, nil)
	err := model.UpdateTitle(c, -1, "")
	require.Error(t, err)
}

func TestUpdateYMD(t *testing.T) {
	t.Parallel()

	c := echo.New().NewContext(nil, nil)
	empty := null.Int16{}
	err := model.UpdateYMD(c, -1, empty, empty, empty)
	require.Error(t, err)

	y := null.Int16From(1900)
	err = model.UpdateYMD(c, -1, y, empty, empty)
	require.ErrorIs(t, err, model.ErrYear)

	m := null.Int16From(13)
	err = model.UpdateYMD(c, -1, empty, m, empty)
	require.ErrorIs(t, err, model.ErrMonth)

	d := null.Int16From(999)
	err = model.UpdateYMD(c, -1, empty, empty, d)
	require.ErrorIs(t, err, model.ErrDay)
}
