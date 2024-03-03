package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOneRecord(t *testing.T) {
	t.Parallel()
	mf, err := model.FindUUID("")
	require.Error(t, err)
	assert.Nil(t, mf)

	mf, err = model.FindUUID("")
	require.Error(t, err)
	assert.Nil(t, mf)

	errID := helper.ObfuscateID(-1)
	mf, err = model.FindUUID(errID)
	require.ErrorIs(t, err, model.ErrID)
	assert.Nil(t, mf)

	errID = helper.ObfuscateID(1)
	mf, err = model.FindUUID(errID)
	require.ErrorIs(t, err, model.ErrDB)
	assert.Nil(t, mf)
}

func TestRecord(t *testing.T) {
	t.Parallel()
	mf, err := model.EditFind(0)
	require.Error(t, err)
	assert.Nil(t, mf)

	mf, err = model.EditFind(0)
	require.Error(t, err)
	assert.Nil(t, mf)

	mf, err = model.EditFind(-1)
	require.ErrorIs(t, err, model.ErrDB)
	assert.Nil(t, mf)

	mf, err = model.EditFind(1)
	require.ErrorIs(t, err, model.ErrDB)
	assert.Nil(t, mf)
}
