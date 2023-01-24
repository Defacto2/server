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

// SoftwareCount counts the number of files that are considered to be software.
func SoftwareCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Soft]; c > 0 {
		return int(c), nil
	}
	c, err := models.Files(SoftwareExpr()).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Soft] = Count(c)
	return int(c), nil
}

// SoftwareByteCount sums the byte filesizes for all the files that are considered to be software.
func SoftwareByteCount(ctx context.Context, db *sql.DB) (int64, error) {
	stmt := "SELECT SUM(files.filesize) AS size_sum FROM files WHERE " +
		fmt.Sprintf("platform = '%s'", tags.Java) +
		fmt.Sprintf("OR platform = '%s'", tags.Linux) +
		fmt.Sprintf("OR platform = '%s'", tags.DOS) +
		fmt.Sprintf("OR platform = '%s'", tags.PHP) +
		fmt.Sprintf("OR platform = '%s'", tags.Windows)
	return models.Files(qm.SQL(stmt)).Count(ctx, db)
}

// SoftwareExpr is a the query mod expression for software files.
func SoftwareExpr() qm.QueryMod {
	java := null.String{String: tags.URIs[tags.Java], Valid: true}
	linux := null.String{String: tags.URIs[tags.Linux], Valid: true}
	dos := null.String{String: tags.URIs[tags.DOS], Valid: true}
	php := null.String{String: tags.URIs[tags.PHP], Valid: true}
	windows := null.String{String: tags.URIs[tags.Windows], Valid: true}
	return qm.Expr(
		models.FileWhere.Platform.EQ(java),
		qm.Or2(models.FileWhere.Platform.EQ(linux)),
		qm.Or2(models.FileWhere.Platform.EQ(dos)),
		qm.Or2(models.FileWhere.Platform.EQ(php)),
		qm.Or2(models.FileWhere.Platform.EQ(windows)),
	)
}
