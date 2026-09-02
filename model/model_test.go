// Package model_test requires an active database connection.
package model_test

import (
	"testing"

	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

// checked in Aug 26, test with an active database, coverage was good at 60%+

func TestValidSceners(t *testing.T) {
	t.Parallel()

	sceners := ""
	r := model.ValidSceners(sceners)
	be.True(t, !r.Valid)

	sceners = "defacto"
	r = model.ValidSceners(sceners)
	be.True(t, r.Valid)
	be.Equal(t, "Defacto", r.String)

	sceners = "defacto, scener    , another person"
	r = model.ValidSceners(sceners)
	be.True(t, r.Valid)
	be.Equal(t, "Defacto,Scener,Another Person", r.String)

	sceners = "dëfå¢T0!"
	r = model.ValidSceners(sceners)
	be.True(t, r.Valid)
	be.Equal(t, "Dëfåt0", r.String)
}
