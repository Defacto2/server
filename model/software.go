package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/Defacto2/server/pkg/tags"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Softs contain statistics for releases that could be considered software.
type Softs struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases that could be considered as digital or pixel art.
func (s *Softs) Stat(ctx context.Context, db *sql.DB) error {
	if s.Bytes > 0 && s.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(SumSize, Counter),
		SoftwareExpr(),
		qm.From(From)).Bind(ctx, db, s)
}

// SoftwareExpr is a the query mod expression for software files.
func SoftwareExpr() qm.QueryMod {
	java := null.String{String: tags.URIs()[tags.Java], Valid: true}
	linux := null.String{String: tags.URIs()[tags.Linux], Valid: true}
	dos := null.String{String: tags.URIs()[tags.DOS], Valid: true}
	php := null.String{String: tags.URIs()[tags.PHP], Valid: true}
	windows := null.String{String: tags.URIs()[tags.Windows], Valid: true}
	return qm.Expr(
		models.FileWhere.Platform.EQ(java),
		qm.Or2(models.FileWhere.Platform.EQ(linux)),
		qm.Or2(models.FileWhere.Platform.EQ(dos)),
		qm.Or2(models.FileWhere.Platform.EQ(php)),
		qm.Or2(models.FileWhere.Platform.EQ(windows)),
	)
}
