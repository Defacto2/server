// Package model_test requires an active database connection.
package model_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	assert.Equal(t, uint64(0), i)
	size = "100"
	i, err = model.ValidFilesize(size)
	require.NoError(t, err)
	assert.Equal(t, uint64(100), i)
	size = "-100"
	i, err = model.ValidFilesize(size)
	require.Error(t, err)
	assert.Equal(t, uint64(0), i)
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

func TestValidPlatform(t *testing.T) {
	t.Parallel()
	tag := ""
	r := model.ValidPlatform(tag)
	assert.False(t, r.Valid)

	tag = "100"
	r = model.ValidPlatform(tag)
	assert.False(t, r.Valid)

	tag = "bbs"
	r = model.ValidPlatform(tag)
	assert.False(t, r.Valid)

	tag = "Windows"
	r = model.ValidPlatform(tag)
	assert.True(t, r.Valid)
	assert.Equal(t, "windows", r.String)
}

func TestValidReleasers(t *testing.T) {
	t.Parallel()
	s1, s2 := "", ""
	r1, r2 := model.ValidReleasers(s1, s2)
	assert.False(t, r1.Valid)
	assert.False(t, r2.Valid)

	s1, s2 = "defacto2", "scene"
	r1, r2 = model.ValidReleasers(s1, s2)
	assert.True(t, r1.Valid)
	assert.True(t, r2.Valid)
	assert.Equal(t, "DEFACTO2", r1.String)
	assert.Equal(t, "SCENE", r2.String)

	// test the swapping of empty releasers
	r1, r2 = model.ValidReleasers("", "defacto2")
	assert.True(t, r1.Valid)
	assert.False(t, r2.Valid)
	assert.Equal(t, "DEFACTO2", r1.String)
	assert.Empty(t, r2.String)
}

func TestValidSceners(t *testing.T) {
	t.Parallel()
	sceners := ""
	r := model.ValidSceners(sceners)
	assert.False(t, r.Valid)

	sceners = "defacto"
	r = model.ValidSceners(sceners)
	assert.True(t, r.Valid)
	assert.Equal(t, "Defacto", r.String)

	sceners = "defacto, scener    , another person"
	r = model.ValidSceners(sceners)
	assert.True(t, r.Valid)
	assert.Equal(t, "Defacto,Scener,Another Person", r.String)

	sceners = "dëfå¢T0!"
	r = model.ValidSceners(sceners)
	assert.True(t, r.Valid)
	assert.Equal(t, "Dëfåt0", r.String)
}

func TestValidSection(t *testing.T) {
	t.Parallel()
	tag := ""
	r := model.ValidSection(tag)
	assert.False(t, r.Valid)

	tag = "100"
	r = model.ValidSection(tag)
	assert.False(t, r.Valid)

	tag = "windows"
	r = model.ValidSection(tag)
	assert.False(t, r.Valid)

	tag = "BBS"
	r = model.ValidSection(tag)
	assert.True(t, r.Valid)
	assert.Equal(t, "bbs", r.String)
}

func TestValidString(t *testing.T) {
	t.Parallel()
	s := "\n\r   \n"
	r := model.ValidString(s)
	assert.False(t, r.Valid)

	const nbsp = "\u00A0"
	r = model.ValidString(nbsp)
	assert.False(t, r.Valid)

	s = "hello world"
	r = model.ValidString(s)
	assert.True(t, r.Valid)
	assert.Equal(t, s, r.String)

	const emoji = "😃"
	r = model.ValidString(emoji)
	assert.True(t, r.Valid)
	assert.Equal(t, emoji, r.String)
}

func TestValidTitle(t *testing.T) {
	t.Parallel()
	title := ""
	r := model.ValidTitle(title)
	assert.False(t, r.Valid)

	title = "hello world"
	r = model.ValidTitle(title)
	assert.True(t, r.Valid)
	assert.Equal(t, title, r.String)

	title = strings.Repeat("a", model.ShortLimit+100)
	r = model.ValidTitle(title)
	assert.True(t, r.Valid)
	assert.Len(t, r.String, model.ShortLimit)
}

func TestValidYouTube(t *testing.T) {
	t.Parallel()
	yt := ""
	r, err := model.ValidYouTube(yt)
	require.NoError(t, err)
	assert.False(t, r.Valid)

	yt = strings.Repeat("x", model.ShortLimit+10)
	r, err = model.ValidYouTube(yt)
	require.NoError(t, err)
	assert.False(t, r.Valid)

	const invalid = "$6BuDfBIcM!"
	r, err = model.ValidYouTube(invalid)
	require.NoError(t, err)
	assert.False(t, r.Valid)

	const valid = "62BuDfBIcMo"
	r, err = model.ValidYouTube(valid)
	require.NoError(t, err)
	assert.True(t, r.Valid)
}

func TestValidNewV7(t *testing.T) {
	t.Parallel()
	now1, unid, err := model.NewV7()
	require.NoError(t, err)

	now2 := time.Now()
	diff := now2.Sub(now1).Minutes()
	const oneMinute = 1.0
	assert.LessOrEqual(t, diff, oneMinute)

	err = uuid.Validate(unid.String())
	require.NoError(t, err)
}
