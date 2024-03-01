package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
)

func TestPublishedFmt(t *testing.T) {
	t.Parallel()
	s := model.PublishedFmt(nil)
	assert.Equal(t, "error, no file model", s)

	ms := models.File{}
	s = model.PublishedFmt(&ms)
	assert.Equal(t, "unknown date", s)

	ms.DateIssuedYear = null.Int16From(1999)
	ms.DateIssuedMonth = null.Int16From(12)
	s = model.PublishedFmt(&ms)
	assert.Equal(t, "1999 December", s)
}
