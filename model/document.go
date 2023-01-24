package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/Defacto2/server/pkg/tags"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// DocumentByteCount sums the byte filesizes for all the files that are considered to be documents.
func DocumentByteCount(ctx context.Context, db *sql.DB) (int64, error) {
	stmt := "SELECT SUM(files.filesize) AS size_sum FROM files WHERE " +
		fmt.Sprintf("platform = '%s'", tags.ANSI) +
		fmt.Sprintf("OR platform = '%s'", tags.Text) +
		fmt.Sprintf("OR platform = '%s'", tags.TextAmiga) +
		fmt.Sprintf("OR platform = '%s'", tags.PDF)
	return models.Files(qm.SQL(stmt)).Count(ctx, db)
}

// DocumentCount counts the number of files that are considered to be documents.
func DocumentCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Doc]; c > 0 {
		return int(c), nil
	}
	c, err := models.Files(DocumentExpr()).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Doc] = Count(c)
	return int(c), nil
}

// DocumentExpr is a the query mod expression for document files.
func DocumentExpr() qm.QueryMod {
	ansi := null.String{String: tags.URIs[tags.ANSI], Valid: true}
	text := null.String{String: tags.URIs[tags.Text], Valid: true}
	amiga := null.String{String: tags.URIs[tags.TextAmiga], Valid: true}
	pdf := null.String{String: tags.URIs[tags.PDF], Valid: true}
	return qm.Expr(
		models.FileWhere.Platform.EQ(ansi),
		qm.Or2(models.FileWhere.Platform.EQ(text)),
		qm.Or2(models.FileWhere.Platform.EQ(amiga)),
		qm.Or2(models.FileWhere.Platform.EQ(pdf)),
	)
}
