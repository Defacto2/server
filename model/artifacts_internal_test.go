package model

import (
	"strings"
	"testing"
	"uuid"

	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/nalgeon/be"
)

func TestOnlyDescriptionsQuery(t *testing.T) {
	t.Parallel()

	be.True(t, onlyDescriptions(nil) == nil)
	be.True(t, onlyDescriptions([]string{" ", ""}) == nil)

	mods := onlyDescriptions([]string{"ultima 7", "ultima 8"})
	sql, args := queries.BuildQuery(models.Files(mods...).Query)

	be.True(t, strings.Contains(sql, "WHERE deletedat IS NULL AND (to_tsvector"))
	be.True(t, strings.Contains(sql, ") AND "))
	be.True(t, strings.Contains(sql, "plainto_tsquery('english'"))
	be.True(t, strings.Contains(sql, "ORDER BY filename ASC"))
	be.Equal(t, args, []any{"ultima 7", "ultima 8"})
}

func TestOnlyFilenamesQuery(t *testing.T) {
	t.Parallel()

	be.True(t, onlyFilenames(nil) == nil)
	be.True(t, onlyFilenames([]string{" ", ""}) == nil)

	mods := onlyFilenames([]string{".iff", "file_id[1].diz", "100%"})
	sql, args := queries.BuildQuery(models.Files(mods...).Query)

	be.True(t, strings.Contains(sql, "WHERE deletedat IS NULL AND (filename ILIKE"))
	be.True(t, strings.Contains(sql, `ESCAPE '\'`))
	be.True(t, strings.Contains(sql, ") AND "))
	be.True(t, strings.Contains(sql, "ORDER BY filename ASC"))
	be.Equal(t, args, []any{"%.iff%", `%file\_id[1].diz%`, `%100\%%`})
}

func TestByFilenameWhere(t *testing.T) {
	t.Parallel()

	clause, args := byFilenameWhere([]string{" ", ""})
	be.Equal(t, clause, "")
	be.True(t, args == nil)

	clause, args = byFilenameWhere([]string{".GIF", "a_b[1]"})
	be.Equal(t, clause, `(filename ILIKE $1 ESCAPE '\' OR filename ILIKE $2 ESCAPE '\') AND deletedat IS NULL`)
	be.Equal(t, args, []any{"%.GIF%", `%a\_b[1]%`})
}

func TestOnlyUniqueIDsQuery(t *testing.T) {
	t.Parallel()

	be.True(t, onlyUniqueIDs(nil) == nil)
	be.True(t, onlyUniqueIDs([]int{0, -1}, uuid.Nil()) == nil)

	uid := uuid.MustParse("c8cd0b4c-2f54-11e0-8827-cc1607e15609")
	mods := onlyUniqueIDs([]int{0, 42, 99}, uuid.Nil(), uid)
	sql, args := queries.BuildQuery(models.Files(mods...).Query)

	be.True(t, strings.Contains(sql, `WHERE ("files"."id" IN ($1,$2)) OR ("files"."uuid" IN ($3))`))
	be.Equal(t, args, []any{int64(42), int64(99), uid.String()})
}
