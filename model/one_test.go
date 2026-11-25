package model_test

import (
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestOne(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	fs, err := model.One(ctx, nil, false, 0)
	be.Err(t, err)
	be.Equal(t, fs, nil)
	_, err = model.OneByUUID(ctx, nil, false, "")
	be.Err(t, err)
	_, err = model.OneFile(ctx, nil, 1)
	be.Err(t, err)

	db := testutil.DB(t)
	_, err = model.One(ctx, db, false, 0)
	be.Err(t, err)
	_, err = model.OneByUUID(ctx, db, false, "")
	be.Err(t, err)

	const id = 1
	fs, err = model.One(ctx, db, false, id)
	be.Err(t, err, nil)
	be.Equal(t, fs.ID, id)
	be.True(t, fs.Filename.String != "")

	uid := fs.UUID.String
	fs1, err := model.OneByUUID(ctx, db, false, uid)
	be.Err(t, err, nil)
	be.Equal(t, fs1.ID, id)

	fs2, err := model.OneFile(ctx, db, id)
	be.Err(t, err, nil)
	be.Equal(t, fs2.ID, id)
}

func TestProdID(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	_, _, err := model.OneDemozoo(ctx, nil, 1)
	be.Err(t, err)
	_, _, err = model.OnePouet(ctx, nil, 1)
	be.Err(t, err)
	_, err = model.OneEditByKey(ctx, nil, "x")
	be.Err(t, err)

	db := testutil.DB(t)

	_, err = model.OneEditByKey(ctx, db, "")
	be.Err(t, err)

	del, key, err := model.OneDemozoo(ctx, db, 0)
	be.Err(t, err, nil)
	be.True(t, !del)
	be.Equal(t, key, 0)

	del, key, err = model.OnePouet(ctx, db, 0)
	be.Err(t, err, nil)
	be.True(t, !del)
	be.Equal(t, key, 0)

	const sharedID = 4148
	const demozoo = 77
	_, key, err = model.OneDemozoo(ctx, db, demozoo)
	be.Err(t, err, nil)
	be.Equal(t, key, sharedID)

	const pouet = 9108
	_, key, err = model.OnePouet(ctx, db, pouet)
	be.Err(t, err, nil)
	be.Equal(t, key, sharedID)

	obf := helper.ObfuscateID(sharedID)
	fs, err := model.OneEditByKey(ctx, db, obf)
	be.Err(t, err, nil)
	be.Equal(t, fs.ID, sharedID)
}
