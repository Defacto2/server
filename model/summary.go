package model

// Package file summary.go contains the database queries for the statistics of files.

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	namer "github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Summary counts the total number files, file sizes and the earliest and latest years.
type Summary struct {
	SumBytes sql.NullInt64 `boil:"size_total"`  // Sum total of the file sizes.
	SumCount sql.NullInt64 `boil:"count_total"` // Sum total count of the files.
	MinYear  sql.NullInt16 `boil:"min_year"`    // Minimum or earliest year of the files.
	MaxYear  sql.NullInt16 `boil:"max_year"`    // Maximum or latest year of the files.
}

// ByDescription saves the summary statistics for the file description search.
func (s *Summary) ByDescription(ctx context.Context, db *sql.DB, terms []string) error {
	if db == nil {
		return ErrDB
	}
	sum := string(postgres.Summary()) // TODO: confirm ClauseNoSoftDel is required.
	for i := range terms {
		const clauseT = "to_tsvector('english', concat_ws(' ', files.record_title, files.comment)) @@ to_tsquery"
		if i == 0 {
			sum = fmt.Sprintf("%s%s($%d) ", sum, clauseT, i+1)
			continue
		}
		sum = fmt.Sprintf("%sOR %s($%d) ", sum, clauseT, i+1)
	}
	sum = strings.TrimSpace(sum)
	return queries.Raw(sum, "'"+strings.Join(terms, "','")+"'").Bind(ctx, db, s)
}

// ByFilename saves the summary statistics for the filename search.
func (s *Summary) ByFilename(ctx context.Context, db *sql.DB, terms []string) error {
	if db == nil {
		return ErrDB
	}
	sum := string(postgres.Summary())
	for i, term := range terms {
		if i == 0 {
			sum += fmt.Sprintf(" filename ~ '%s' OR filename ILIKE '%s' OR filename ILIKE '%s' OR filename ILIKE '%s'",
				term, term+"%", "%"+term, "%"+term+"%")
			continue
		}
		sum += fmt.Sprintf(" OR filename ~ '%s' OR filename ILIKE '%s' OR filename ILIKE '%s' OR filename ILIKE '%s'",
			term, term+"%", "%"+term, "%"+term+"%")
	}
	sum = strings.TrimSpace(sum)
	return queries.Raw(sum).Bind(ctx, db, s)
}

// ByForApproval returns the summary statistics for files that require approval.
func (s *Summary) ByForApproval(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	boil.DebugMode = true
	return models.NewQuery(
		models.FileWhere.Deletedat.IsNotNull(),
		qm.WithDeleted(),
		qm.Select(postgres.Columns()...),
		qm.From(From)).Bind(ctx, db, s)
}

// ByHidden returns the summary statistics for files that have been deleted.
func (s *Summary) ByHidden(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	boil.DebugMode = true
	return models.NewQuery(
		models.FileWhere.Deletedat.IsNotNull(),
		models.FileWhere.Deletedby.IsNotNull(),
		qm.WithDeleted(),
		qm.Select(postgres.Columns()...),
		qm.From(From)).Bind(ctx, db, s)
}

// ByPublic selects the summary statistics for all public files.
func (s *Summary) ByPublic(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		qm.From(From)).Bind(ctx, db, s)
}

// ByScener selects the summary statistics for the named sceners.
func (s *Summary) ByScener(ctx context.Context, db *sql.DB, name string) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(postgres.ScenerSQL(name)),
		qm.Where(ClauseNoSoftDel),
		qm.From(From)).Bind(ctx, db, s)
}

// ByReleaser returns the summary statistics for the named releaser.
// The name is case insensitive and should be the URI slug of the releaser.
func (s *Summary) ByReleaser(ctx context.Context, db *sql.DB, name string) error {
	if db == nil {
		return ErrDB
	}
	ns, err := namer.Humanize(namer.Path(name))
	if err != nil {
		return fmt.Errorf("namer.Humanize: %w", err)
	}
	n := strings.ToUpper(ns)
	x := null.StringFrom(n)
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where("upper(group_brand_for) = ? OR upper(group_brand_by) = ?", x, x),
		qm.Where(ClauseNoSoftDel),
		qm.From(From)).Bind(ctx, db, s)
}

// ByUnwanted returns the summary statistics for files that have been marked as unwanted.
func (s *Summary) ByUnwanted(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	boil.DebugMode = true
	return models.NewQuery(
		models.FileWhere.FileSecurityAlertURL.IsNotNull(),
		qm.WithDeleted(),
		qm.Select(postgres.Columns()...),
		qm.From(From)).Bind(ctx, db, s)
}

// ByMatch returns the summary statistics for the named uri.
func (s *Summary) ByMatch(ctx context.Context, db *sql.DB, uri string) error { //nolint:lll,funlen,gocognit,gocyclo,cyclop,maintidx
	if db == nil {
		return ErrDB
	}
	var c, b, y0, y1 int
	var err error
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
		c, b, y0, y1, err = textAmiga(ctx, db)
	case "text-apple2":
		c, b, y0, y1, err = textApple2(ctx, db)
	case "text-atari-st":
		c, b, y0, y1, err = textAtariST(ctx, db)
	case "pdf":
		c, b, y0, y1, err = pdf(ctx, db)
	case "html":
		c, b, y0, y1, err = html(ctx, db)
	case "news-article":
		c, b, y0, y1, err = newsArticle(ctx, db)
	case "standards":
		c, b, y0, y1, err = standards(ctx, db)
	case "announcement":
		c, b, y0, y1, err = announcement(ctx, db)
	case "job-advert":
		c, b, y0, y1, err = jobAdvert(ctx, db)
	case "trial-crackme":
		c, b, y0, y1, err = trialCrackme(ctx, db)
	case "hack":
		c, b, y0, y1, err = hack(ctx, db)
	case "tool":
		c, b, y0, y1, err = tool(ctx, db)
	case "takedown":
		c, b, y0, y1, err = takedown(ctx, db)
	case "drama":
		c, b, y0, y1, err = drama(ctx, db)
	case "advert":
		c, b, y0, y1, err = advert(ctx, db)
	case "restrict":
		c, b, y0, y1, err = restrict(ctx, db)
	case "how-to":
		c, b, y0, y1, err = howTo(ctx, db)
	case "nfo-tool":
		c, b, y0, y1, err = nfoTool(ctx, db)
	case "image":
		c, b, y0, y1, err = image(ctx, db)
	case "music":
		c, b, y0, y1, err = music(ctx, db)
	case "video":
		c, b, y0, y1, err = video(ctx, db)
	case "msdos":
		c, b, y0, y1, err = msdos(ctx, db)
	case "windows":
		c, b, y0, y1, err = windows(ctx, db)
	case "macos":
		c, b, y0, y1, err = macos(ctx, db)
	case "linux":
		c, b, y0, y1, err = linux(ctx, db)
	case "java":
		c, b, y0, y1, err = java(ctx, db)
	case "script":
		c, b, y0, y1, err = script(ctx, db)
	default:
		return fmt.Errorf("%w: %q", ErrURI, uri)
	}
	if err != nil {
		return err
	}
	s.SumBytes = sql.NullInt64{Int64: int64(b)}
	s.SumCount = sql.NullInt64{Int64: int64(c)}
	s.MinYear = sql.NullInt16{Int16: int16(y0)}
	s.MaxYear = sql.NullInt16{Int16: int16(y1)}
	return nil
}

func textAmiga(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := TextAmiga{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("textAmiga.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func textApple2(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := TextApple2{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("textApple2.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func textAtariST(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := TextAtariST{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("textAtariST.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func pdf(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := PDF{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("pdf.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func html(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := HTML{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("html.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func newsArticle(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := NewsArticle{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("newsArticle.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func standards(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Standard{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("standards.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func announcement(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Announcement{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("announcement.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func jobAdvert(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := JobAdvert{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("jobAdvert.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func trialCrackme(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := TrialCrackme{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("trailCrackme.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func hack(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Hack{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("hack.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func tool(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Tool{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("tool.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func takedown(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Takedown{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("takedown.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func drama(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Drama{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("drama.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func advert(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Advert{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("advert.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func restrict(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Restrict{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("restrict.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func howTo(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := HowTo{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("howTo.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func nfoTool(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := NfoTool{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("nfoTool.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func image(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Image{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("image.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func music(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Music{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("music.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func video(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Video{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("video.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func msdos(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := MsDos{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("msdos.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func windows(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Windows{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("windows.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func macos(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Macos{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("macos.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func linux(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Linux{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("linux.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func java(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Java{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("java.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}

func script(ctx context.Context, db *sql.DB) (int, int, int, int, error) {
	m := Script{}
	if err := m.Stat(ctx, db); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("script.Stat: %w", err)
	}
	return m.Count, m.Bytes, m.MinYear, m.MaxYear, nil
}
