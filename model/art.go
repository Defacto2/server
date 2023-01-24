package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/Defacto2/server/tags"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// ArtCount counts the files that could be considered as digital or pixel art.
func ArtCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Art]; c > 0 {
		return int(c), nil
	}
	c, err := models.Files(ArtExpr()).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Art] = Count(c)
	return int(c), nil
}

// ArtByteCount sums the byte filesizes for all the files that is considered as digital or pixel art.
func ArtByteCount(ctx context.Context, db *sql.DB) (int64, error) {
	stmt := "SELECT SUM(files.filesize) AS size_sum FROM files WHERE" +
		fmt.Sprintf(" files.section != '%s'", tags.BBS) +
		fmt.Sprintf(" AND files.platform = '%s';", tags.Image)
	return models.Files(qm.SQL(stmt)).Count(ctx, db)
}

// ArtExpr is a the query mod expression for art files.
func ArtExpr() qm.QueryMod {
	bbs := null.String{String: tags.URIs[tags.BBS], Valid: true}
	image := null.String{String: tags.URIs[tags.Image], Valid: true}
	return qm.Expr(
		models.FileWhere.Section.NEQ(bbs),
		models.FileWhere.Platform.EQ(image),
	)
}
