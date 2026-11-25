package model_test

import (
	"encoding/hex"
	"strconv"
	"strings"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestExist(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	// Offline nil-executor guards
	_, err := model.ExistDemozoo(ctx, nil, 1)
	be.Err(t, err)
	_, err = model.ExistPouet(ctx, nil, 1)
	be.Err(t, err)
	_, err = model.ExistSHA(ctx, nil, nil)
	be.Err(t, err)
	_, err = model.ExistHash(ctx, nil, "")
	be.Err(t, err)
	_, err = model.OneByHash(ctx, nil, "")
	be.Err(t, err)
	_, err = model.ExistUUID(ctx, nil, "")
	be.Err(t, err)

	db := testutil.DB(t)

	const (
		key     = 108
		missing = 999_999_999
	)
	ok, err := model.ExistDemozoo(ctx, db, key)
	be.Err(t, err, nil)
	be.True(t, ok)

	ok, err = model.ExistDemozoo(ctx, db, missing)
	be.Err(t, err, nil)
	be.True(t, !ok)

	ok, err = model.ExistPouet(ctx, db, missing)
	be.Err(t, err, nil)
	be.True(t, !ok)

	ok, err = model.ExistSHA(ctx, db, nil)
	be.Err(t, err, nil)
	be.True(t, !ok)

	const hexd = "12a32a20535db22f3a39f1dde0e96b1534f9cb6643f48ac73dcd44f84dd55aadd7e0711b222931ae8d3df89828372416"
	decd, err := hex.DecodeString(hexd)
	be.Err(t, err, nil)

	ok, err = model.ExistSHA(ctx, db, decd)
	be.Err(t, err, nil)
	be.True(t, ok)

	_, err = model.ExistHash(ctx, db, "")
	be.Err(t, err, model.ErrSha384)

	zeroHash := strings.Repeat("0", 96)
	ok, err = model.ExistHash(ctx, db, zeroHash)
	be.Err(t, err, nil)
	be.True(t, !ok)

	ok, err = model.ExistHash(ctx, db, hexd)
	be.Err(t, err, nil)
	be.True(t, ok)

	_, err = model.OneByHash(ctx, db, "")
	be.Err(t, err, model.ErrSha384)

	obf, err := model.OneByHash(ctx, db, zeroHash)
	be.Err(t, err, nil)
	be.Equal(t, obf, "")

	obf, err = model.OneByHash(ctx, db, hexd)
	be.Err(t, err, nil)
	be.Equal(t, obf, "9b1c6")

	idkey := helper.DeObfuscate(obf)
	id, err := strconv.Atoi(idkey)
	be.Err(t, err, nil)
	be.Equal(t, id, 1)

	ok, err = model.ExistUUID(ctx, db, "")
	be.Err(t, err, nil)
	be.True(t, !ok)

	const uid = "c8cd0b4c-2f54-11e0-8827-cc1607e15609"
	ok, err = model.ExistUUID(ctx, db, uid)
	be.Err(t, err, nil)
	be.True(t, ok)
}
