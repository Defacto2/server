// Package model_test requires an active database connection.
package model_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/google/uuid"
	"github.com/nalgeon/be"
)

// checked in Aug 26, test coverage was poor at around 8%~

func openDB(t *testing.T) *sql.DB {
	db, err := postgres.Open()
	if err != nil {
		t.Log("postgres open", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		return nil
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Log("cleanup database", err)
		}
	})

	return db
}

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

func TestValidNewV7(t *testing.T) {
	t.Parallel()

	now1, unid, err := model.NewV7()
	be.Err(t, err, nil)

	now2 := time.Now()
	diff := now2.Sub(now1).Minutes()

	const oneMinute = 1.0
	be.True(t, diff <= oneMinute)
	err = uuid.Validate(unid.String())
	be.Err(t, err, nil)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	err := model.DeleteOne(t.Context(), nil, -1)
	be.Err(t, err)
}

func TestModel(t *testing.T) {
	t.Parallel()

	_, err := model.JsDosBinary(nil)
	be.Err(t, err)

	_, err = model.JsDosConfig(nil)
	be.Err(t, err)

	_, err = model.JsDosCommand(nil)
	be.Err(t, err)
}
