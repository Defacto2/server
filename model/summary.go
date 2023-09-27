package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Defacto2/sceners/pkg/rename"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Package file summary.go contains the database queries for the statistics of files.

// Summary counts the total number files, file sizes and the earliest and latest years.
type Summary struct {
	SumBytes int `boil:"size_total"`  // Sum total of the file sizes.
	SumCount int `boil:"count_total"` // Sum total count of the files.
	MinYear  int `boil:"min_year"`    // Minimum or earliest year of the files.
	MaxYear  int `boil:"max_year"`    // Maximum or latest year of the files.
}

func (r *Summary) Search(ctx context.Context, db *sql.DB, terms []string) error {
	if db == nil {
		return ErrDB
	}
	s := "SELECT COUNT(files.id) AS count_total, " +
		"SUM(files.filesize) AS size_total, " +
		"MIN(files.date_issued_year) AS min_year, " +
		"MAX(files.date_issued_year) AS max_year " +
		"FROM files " +
		"WHERE "
	for i, term := range terms {
		if i > 0 {
			s = fmt.Sprintf("%s OR ", s)
		}
		s = fmt.Sprintf("%s files.filename ~ '%s' OR files.filename ILIKE '%s' ", s, term, "%"+term+"%")
	}
	s = strings.TrimSpace(s)
	fmt.Println(s)
	if err := queries.Raw(s).Bind(ctx, db, r); err != nil {
		return err
	}
	return nil
}

func (s *Summary) All(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		qm.From(From)).Bind(ctx, db, s)
}

func (r *Summary) BBS(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if err := queries.Raw(string(postgres.SumBBS())).Bind(ctx, db, r); err != nil {
		return err
	}
	return nil
}

func (r *Summary) FTP(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if err := queries.Raw(string(postgres.SumFTP())).Bind(ctx, db, r); err != nil {
		return err
	}
	return nil
}

func (r *Summary) Magazine(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if err := queries.Raw(string(postgres.SumMag())).Bind(ctx, db, r); err != nil {
		return err
	}
	return nil
}

func (s *Summary) Scener(ctx context.Context, db *sql.DB, name string) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ScenerSQL(name)),
		qm.From(From)).Bind(ctx, db, s)
}

// Releaser returns the summary statistics for the named releaser.
// The name is case insensitive and should be the URI slug of the releaser.
func (s *Summary) Releaser(ctx context.Context, db *sql.DB, name string) error {
	if db == nil {
		return ErrDB
	}
	n := strings.ToUpper(rename.DeObfuscateURL(name))
	x := null.StringFrom(n)
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where("upper(group_brand_for) = ? OR upper(group_brand_by) = ?", x, x),
		// qm.Or2(models.FileWhere.Platform.EQ(expr.PText())
		qm.From(From)).Bind(ctx, db, s)
}

func (s *Summary) URI(ctx context.Context, db *sql.DB, uri string) error {
	if db == nil {
		return ErrDB
	}
	c, b, y0, y1 := 0, 0, 0, 0
	switch uri {
	case "intro-windows":
		m := IntroWindows{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "intro-msdos":
		m := IntroMsDos{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "intro":
		m := Intro{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "installer":
		m := Installer{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "demoscene":
		m := Demoscene{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "nfo":
		m := Nfo{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "proof":
		m := Proof{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "ansi":
		m := Ansi{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "ansi-brand":
		m := AnsiBrand{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "ansi-bbs":
		m := AnsiBBS{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "ansi-ftp":
		m := AnsiFTP{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "ansi-pack":
		m := AnsiPack{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "ansi-nfo":
		m := AnsiNfo{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "bbs":
		m := BBS{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "bbstro":
		m := BBStro{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "bbs-image":
		m := BBSImage{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "bbs-text":
		m := BBSText{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "ftp":
		m := FTP{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "magazine":
		m := Magazine{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "text":
		m := Text{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "text-pack":
		m := TextPack{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "image-pack":
		m := ImagePack{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "windows-pack":
		m := WindowsPack{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "msdos-pack":
		m := MsDosPack{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "database":
		m := Database{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "text-amiga":
		m := TextAmiga{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "text-apple2":
		m := TextApple2{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "text-atari-st":
		m := TextAtariST{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "pdf":
		m := PDF{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "html":
		m := HTML{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "news-article":
		m := NewsArticle{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "standards":
		m := Standard{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "announcement":
		m := Announcement{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "job-advert":
		m := JobAdvert{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "trial-crackme":
		m := TrialCrackme{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "hack":
		m := Hack{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "tool":
		m := Tool{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "takedown":
		m := Takedown{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "drama":
		m := Drama{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "advert":
		m := Advert{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "restrict":
		m := Restrict{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "how-to":
		m := HowTo{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "nfo-tool":
		m := NfoTool{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "image":
		m := Image{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "music":
		m := Music{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "video":
		m := Video{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "msdos":
		m := MsDos{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "windows":
		m := Windows{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "macos":
		m := Macos{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "linux":
		m := Linux{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "java":
		m := Java{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	case "script":
		m := Script{}
		if err := m.Stat(ctx, db); err != nil {
			return err
		}
		c, b, y0, y1 = m.Count, m.Bytes, m.MinYear, m.MaxYear
	default:
		return fmt.Errorf("%w: %q", ErrURI, uri)
	}
	s.SumBytes, s.SumCount, s.MinYear, s.MaxYear = b, c, y0, y1
	return nil
}
