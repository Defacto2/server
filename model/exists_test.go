package model_test

import (
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestExist(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	const key = 108
	ok, err := model.ExistDemozoo(t.Context(), db, key)
	be.Err(t, err, nil)
	be.True(t, ok)

	_, err = model.ExistSHA(nil, nil, nil)
	be.Err(t, err)

	ok, err = model.ExistSHA(t.Context(), db, nil)
	be.Err(t, err, nil)
	be.True(t, !ok)

	const hexd = "12a32a20535db22f3a39f1dde0e96b1534f9cb6643f48ac73dcd44f84dd55aadd7e0711b222931ae8d3df89828372416"
	decd, err := hex.DecodeString(hexd)
	be.Err(t, err, nil)

	ok, err = model.ExistSHA(t.Context(), db, decd)
	be.Err(t, err, nil)
	be.True(t, ok)

	ok, err = model.ExistHash(t.Context(), db, hexd)
	be.Err(t, err, nil)
	be.True(t, ok)

	_, err = model.OneByHash(t.Context(), db, "")
	be.Err(t, err)

	obf, err := model.OneByHash(t.Context(), db, hexd)
	be.Err(t, err, nil)
	be.Equal(t, obf, "9b1c6")

	idkey := helper.DeObfuscate(obf)
	id, err := strconv.Atoi(idkey)
	be.Err(t, err, nil)
	be.Equal(t, id, 1)

	ok, err = model.ExistUUID(t.Context(), db, "")
	be.Err(t, err, nil)
	be.True(t, !ok)

	const uid = "c8cd0b4c-2f54-11e0-8827-cc1607e15609"
	ok, err = model.ExistUUID(t.Context(), db, uid)
	be.Err(t, err, nil)
	be.True(t, ok)
}
