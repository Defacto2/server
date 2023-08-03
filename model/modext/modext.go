package modext

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func Time(ctx context.Context, db *sql.DB, f *models.File) {
	// placeholder
}

// DemoExpr is a the query mod expression for demo releases.
func DemoExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SDemo()),
	)
}

// InstallExpr is a the query mod expression for install releases.
func InstallExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SInstall()),
	)
}

// IntroExpr is a the query mod expression for intro releases.
func IntroExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SIntro()),
	)
}

// IntroDOSExpr is a the query mod expression for intro releases on MS-DOS.
func IntroDOSExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SIntro()),
		models.FileWhere.Platform.EQ(PDos()),
	)
}

// IntroWindowsExpr is a the query mod expression for Windows intro releases.
func IntroWindowsExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SIntro()),
		models.FileWhere.Platform.EQ(PWindows()),
	)
}

// DOSExpr is a the query mod expression for MS-DOS releases.
func DOSExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PDos()),
	)
}

// LinuxExpr is a the query mod expression for Linux releases.
func LinuxExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PLinux()),
	)
}

// JavaExpr is a the query mod expression for Java releases.
func JavaExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PJava()),
	)
}

// MacExpr is a the query mod expression for MacOS releases.
func MacExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PMac()),
	)
}

// ScriptExpr is a the query mod expression for script and shell releases.
func ScriptExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PScript()),
	)
}

// WindowsExpr is a the query mod expression for Windows releases.
func WindowsExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PWindows()),
	)
}
