package model_test

import (
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestOnes(t *testing.T) {
	t.Parallel()

	fs, got := model.One(nil, nil, false, 0)
	be.Err(t, got)
	be.Equal(t, fs, nil)

	db := testutil.DB(t)
	_, got = model.One(t.Context(), db, false, 0)
	be.Err(t, got)

	const id = 1
	fs, got = model.One(t.Context(), db, false, 1)
	be.Err(t, got, nil)
	be.Equal(t, fs.ID, id)
	be.True(t, fs.Filename.String != "")

	uid := fs.UUID.String
	fs1, got := model.OneByUUID(t.Context(), db, false, uid)
	be.Err(t, got, nil)
	be.Equal(t, fs1.ID, id)

	fs2, got := model.OneFile(t.Context(), db, id)
	be.Err(t, got, nil)
	be.Equal(t, fs2.ID, id)
}

func TestProdID(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	del, key, got := model.OneDemozoo(t.Context(), db, 0)
	be.Err(t, got, nil)
	be.True(t, !del)
	be.Equal(t, key, 0)

	del, key, got = model.OnePouet(t.Context(), db, 0)
	be.Err(t, got, nil)
	be.True(t, !del)
	be.Equal(t, key, 0)

	const keyn = 56097
	const dzblacklotus = 12
	_, key, got = model.OneDemozoo(t.Context(), db, dzblacklotus)
	be.Err(t, got, nil)
	be.Equal(t, key, keyn)

	const ptblacklotus = 2
	_, key, got = model.OnePouet(t.Context(), db, ptblacklotus)
	be.Err(t, got, nil)
	be.Equal(t, key, keyn)

	obf := helper.ObfuscateID(keyn)
	fs, got := model.OneEditByKey(t.Context(), db, obf)
	be.Err(t, got, nil)
	be.Equal(t, fs.ID, keyn)
}
