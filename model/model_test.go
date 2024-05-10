// Package model_test requires an active database connection.
package model_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/model"
	"github.com/stretchr/testify/assert"
)

func TestValidDateIssue(t *testing.T) {
	t.Parallel()
	y, m, d := model.ValidDateIssue("", "", "")
	assert.False(t, y.Valid)
	assert.False(t, m.Valid)
	assert.False(t, d.Valid)
	y, _, _ = model.ValidDateIssue("1980", "", "")
	assert.True(t, y.Valid)
	assert.Equal(t, int16(1980), y.Int16)
	y, m, d = model.ValidDateIssue("9999", "999", "999")
	assert.False(t, y.Valid)
	assert.False(t, m.Valid)
	assert.False(t, d.Valid)
	y, m, d = model.ValidDateIssue("1980", "1", "2")
	assert.True(t, y.Valid)
	assert.Equal(t, int16(1980), y.Int16)
	assert.True(t, m.Valid)
	assert.Equal(t, int16(1), m.Int16)
	assert.True(t, d.Valid)
	assert.Equal(t, int16(2), d.Int16)
}

func TestValidFilename(t *testing.T) {
	t.Parallel()
	name := ""
	r := model.ValidFilename(name)
	assert.False(t, r.Valid)

	name = "somefile.txt"
	r = model.ValidFilename(name)
	assert.True(t, r.Valid)
	assert.Equal(t, name, r.String)

	name = strings.Repeat("a", model.LongFilename+100)
	r = model.ValidFilename(name)
	assert.True(t, r.Valid)
	assert.Len(t, r.String, model.LongFilename)
}

func TestValidFilesize(t *testing.T) {
	t.Parallel()
	size := ""
	i, err := model.ValidFilesize(size)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), i)
	size = "100"
	i, err = model.ValidFilesize(size)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), i)
	size = "-100"
	i, err = model.ValidFilesize(size)
	assert.Error(t, err)
	assert.Equal(t, int64(0), i)
}

func TestValidIntegrity(t *testing.T) {
	t.Parallel()
	integ := ""
	r := model.ValidIntegrity(integ)
	assert.False(t, r.Valid)
	assert.Empty(t, r.String)

	integ = "abcde"
	r = model.ValidIntegrity(integ)
	assert.False(t, r.Valid)
	assert.Empty(t, r.String)

	const valid = "8ac9e700d8d5467fb8f62c88628b1f30cbfa1d0696a81a78599af01bb913cc726a78f3817adfa557691db9ad1354df6b"
	r = model.ValidIntegrity(valid)
	assert.True(t, r.Valid)
	assert.Equal(t, valid, r.String)

	const invalid = "XXXXXX00d8d5467fb8f62c88628b1f30cbfa1d0696a81a78599af01bb913cc726a78f3817adfa557691db9ad1354df6b"
	r = model.ValidIntegrity(invalid)
	assert.False(t, r.Valid)
	assert.Empty(t, r.String)
}

func TestValidLastMod(t *testing.T) {
	t.Parallel()
	lastmod := ""
	r := model.ValidLastMod(lastmod)
	assert.False(t, r.Valid)

	lastmod = "100"
	r = model.ValidLastMod(lastmod)
	assert.False(t, r.Valid)

	oneHourAgo := time.Now().Add(-time.Hour).UnixMilli()
	lastmod = strconv.FormatInt(oneHourAgo, 10)
	r = model.ValidLastMod(lastmod)
	assert.True(t, r.Valid)
	assert.Greater(t, time.Now().UnixMilli(), r.Time.UnixMilli())

	oneHourFromNow := time.Now().Add(time.Hour).UnixMilli()
	lastmod = strconv.FormatInt(oneHourFromNow, 10)
	r = model.ValidLastMod(lastmod)
	assert.False(t, r.Valid)
}

func TestValidMagic(t *testing.T) {
	t.Parallel()
	magic := ""
	r := model.ValidMagic(magic)
	assert.False(t, r.Valid)

	magic = "100"
	r = model.ValidMagic(magic)
	assert.False(t, r.Valid)

	magic = "defacto2"
	r = model.ValidMagic(magic)
	assert.False(t, r.Valid)

	magic = "Text/HTML"
	r = model.ValidMagic(magic)
	assert.True(t, r.Valid)
	assert.Equal(t, "text/html", r.String)
}
