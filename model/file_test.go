package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func logr() *zap.SugaredLogger {
	return zap.NewExample().Sugar()
}

func TestOneRecord(t *testing.T) {
	t.Parallel()
	mf, err := model.OneRecord(nil, nil, false, "")
	assert.Error(t, err)
	assert.Nil(t, mf)

	mf, err = model.OneRecord(logr(), nil, false, "")
	assert.Error(t, err)
	assert.Nil(t, mf)

	c := echo.New().NewContext(nil, nil)

	errID := helper.ObfuscateID(-1)
	mf, err = model.OneRecord(logr(), c, false, errID)
	assert.ErrorIs(t, err, model.ErrID)
	assert.Nil(t, mf)

	errID = helper.ObfuscateID(1)
	mf, err = model.OneRecord(logr(), c, false, errID)
	assert.ErrorIs(t, err, model.ErrDB)
	assert.Nil(t, mf)
}

func TestRecord(t *testing.T) {
	t.Parallel()
	mf, err := model.Record(nil, nil, 0)
	assert.Error(t, err)
	assert.Nil(t, mf)

	mf, err = model.Record(logr(), nil, 0)
	assert.Error(t, err)
	assert.Nil(t, mf)

	c := echo.New().NewContext(nil, nil)

	mf, err = model.Record(logr(), c, -1)
	assert.ErrorIs(t, err, model.ErrDB)
	assert.Nil(t, mf)

	mf, err = model.Record(logr(), c, 1)
	assert.ErrorIs(t, err, model.ErrDB)
	assert.Nil(t, mf)
}
