package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/Defacto2/server/pkg/tags"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Arts contain statistics for releases that could be considered as digital or pixel art.
type Arts struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases that could be considered as digital or pixel art.
func (a *Arts) Stat(ctx context.Context, db *sql.DB) error {
	if a.Bytes > 0 && a.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(SumSize, Counter),
		ArtExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

// ArtExpr is a the query mod expression for art releases.
func ArtExpr() qm.QueryMod {
	bbs := null.String{String: tags.URIs[tags.BBS], Valid: true}
	image := null.String{String: tags.URIs[tags.Image], Valid: true}
	return qm.Expr(
		models.FileWhere.Section.NEQ(bbs),
		models.FileWhere.Platform.EQ(image),
	)
}
