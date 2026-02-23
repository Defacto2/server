package model

// Package file summary.go contains the database queries for the statistics of files.

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"

	namer "github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// Summary counts the total number files, file sizes and the earliest and latest years.
type Summary struct {
	SumBytes sql.NullInt64 `boil:"size_total"`  // Sum total of the file sizes.
	SumCount sql.NullInt64 `boil:"count_total"` // Sum total count of the files.
	MinYear  sql.NullInt16 `boil:"min_year"`    // Minimum or earliest year of the files.
	MaxYear  sql.NullInt16 `boil:"max_year"`    // Maximum or latest year of the files.
}

// ByDescription saves the summary statistics for the file description search.
func (s *Summary) ByDescription(ctx context.Context, exec boil.ContextExecutor, terms []string) error {
	panics.BoilExecCrash(exec)
	sum := string(postgres.Summary())
	for i := range terms {
		const clauseT = "to_tsvector('english', concat_ws(' ', files.record_title, files.comment)) @@ websearch_to_tsquery"
		if i == 0 {
			sum = fmt.Sprintf("%s%s($%d) ", sum, clauseT, i+1)
			continue
		}
		sum = fmt.Sprintf("%sOR %s($%d) ", sum, clauseT, i+1)
	}
	sum += "AND " + ClauseNoSoftDel
	sum = strings.TrimSpace(sum)
	return queries.Raw(sum, "'"+strings.Join(terms, "','")+"'").Bind(ctx, exec, s)
}

// ByFilename saves the summary statistics for the filename search.
func (s *Summary) ByFilename(ctx context.Context, exec boil.ContextExecutor, terms []string) error {
	panics.BoilExecCrash(exec)
	var sum strings.Builder
	sum.WriteString(string(postgres.Summary()))
	for i, term := range terms {
		if i == 0 {
			fmt.Fprintf(&sum, " filename ~ '%s' OR filename ILIKE '%s' OR filename ILIKE '%s' OR filename ILIKE '%s'",
				term, term+"%", "%"+term, "%"+term+"%")
			continue
		}
		fmt.Fprintf(&sum, " OR filename ~ '%s' OR filename ILIKE '%s' OR filename ILIKE '%s' "+
			"OR filename ILIKE '%s'", term, term+"%", "%"+term, "%"+term+"%")
	}
	sum.WriteString("AND " + ClauseNoSoftDel)
	query := strings.TrimSpace(sum.String())
	return queries.Raw(query).Bind(ctx, exec, s)
}

// ByForApproval returns the summary statistics for files that require approval.
func (s *Summary) ByForApproval(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		models.FileWhere.Deletedat.IsNotNull(),
		models.FileWhere.Deletedby.IsNull(),
		qm.WithDeleted(),
		qm.Select(postgres.Columns()...),
		qm.From(From)).Bind(ctx, exec, s)
}

// ByHidden returns the summary statistics for files that have been deleted.
func (s *Summary) ByHidden(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		models.FileWhere.Deletedat.IsNotNull(),
		models.FileWhere.Deletedby.IsNotNull(),
		qm.WithDeleted(),
		qm.Select(postgres.Columns()...),
		qm.From(From)).Bind(ctx, exec, s)
}

// ByPublic selects the summary statistics for all public files.
func (s *Summary) ByPublic(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		qm.From(From)).Bind(ctx, exec, s)
}

// ByScener selects the summary statistics for the named sceners.
func (s *Summary) ByScener(ctx context.Context, exec boil.ContextExecutor, name string) error {
	panics.BoilExecCrash(exec)
	query, params := postgres.ScenerSQL(name)
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(query, params...),
		qm.Where(ClauseNoSoftDel),
		qm.From(From)).Bind(ctx, exec, s)
}

// ByReleaser returns the summary statistics for the named releaser.
// The name is case insensitive and should be the URI slug of the releaser.
func (s *Summary) ByReleaser(ctx context.Context, exec boil.ContextExecutor, name string) error {
	panics.BoilExecCrash(exec)
	ns, err := namer.Humanize(namer.Path(name))
	if err != nil {
		return fmt.Errorf("summary by releaser namer humanize: %w", err)
	}
	n := strings.ToUpper(ns)
	x := null.StringFrom(n)
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where("upper(group_brand_for) = ? OR upper(group_brand_by) = ?", x, x),
		qm.Where(ClauseNoSoftDel),
		qm.From(From)).Bind(ctx, exec, s)
}

// ByUnwanted returns the summary statistics for files that have been marked as unwanted.
func (s *Summary) ByUnwanted(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	empty := null.StringFrom("")
	return models.NewQuery(
		models.FileWhere.FileSecurityAlertURL.IsNotNull(),
		models.FileWhere.FileSecurityAlertURL.NEQ(empty),
		qm.WithDeleted(),
		qm.Select(postgres.Columns()...),
		qm.From(From)).Bind(ctx, exec, s)
}

// Update updates the summary statistics.
func (s *Summary) Update(c, b, y0, y1 int) {
	s.SumCount = sql.NullInt64{Int64: int64(c)}
	s.SumBytes = sql.NullInt64{Int64: int64(b)}
	s.MinYear = sql.NullInt16{Int16: int16(math.Abs(float64(y0)))}
	s.MaxYear = sql.NullInt16{Int16: int16(math.Abs(float64(y1)))}
}

// StatFunc is a function that updates the summary statistics.
type StatFunc func(context.Context, boil.ContextExecutor) error

func (s *Summary) Matches() map[string]StatFunc {
	return map[string]StatFunc{
		"text-amiga":    s.textAmiga,
		"text-apple2":   s.textApple2,
		"text-atari-st": s.textAtariST,
		"pdf":           s.pdf,
		"html":          s.html,
		"news-article":  s.newsArticle,
		"standards":     s.standards,
		"announcement":  s.announcement,
		"job-advert":    s.jobAdvert,
		"trial-crackme": s.trialCrackme,
		"hack":          s.hack,
		"tool":          s.tool,
		"takedown":      s.takedown,
		"drama":         s.drama,
		"advert":        s.advert,
		"restrict":      s.restrict,
		"how-to":        s.howTo,
		"nfo-tool":      s.nfoTool,
		"image":         s.image,
		"music":         s.music,
		"video":         s.video,
		"msdos":         s.msdos,
		"windows":       s.windows,
		"macos":         s.macos,
		"linux":         s.linux,
		"java":          s.java,
		"script":        s.script,
		"database":      s.database,
		"msdos-pack":    s.msdosPack,
		"windows-pack":  s.windowsPack,
		"image-pack":    s.imagePack,
		"text-pack":     s.textPack,
		"text":          s.text,
		"magazine":      s.magazine,
		"ftp":           s.ftp,
		"bbs-text":      s.bbsText,
		"bbs-image":     s.bbsImage,
		"bbstro":        s.bbstro,
		"bbs":           s.bbs,
		"ansi-nfo":      s.ansiNfo,
		"ansi-pack":     s.ansiPack,
		"ansi-ftp":      s.ansiFTP,
		"ansi-bbs":      s.ansiBBS,
		"ansi-brand":    s.ansiBrand,
		"ansi":          s.ansi,
		"proof":         s.proof,
		"nfo":           s.nfo,
		"demoscene":     s.demoscene,
		"installer":     s.installer,
		"intro":         s.intro,
		"intro-msdos":   s.introMsdos,
		"intro-windows": s.introWindows,
	}
}

// ByMatch returns the summary statistics for the named uri.
func (s *Summary) ByMatch(ctx context.Context, exec boil.ContextExecutor, uri string) error {
	stat := s.Matches()
	if update, match := stat[uri]; match {
		return update(ctx, exec)
	}
	return fmt.Errorf("%w: %q", ErrURI, uri)
}

func (s *Summary) introWindows(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := IntroWindows{}
	if err := m.Stat(ctx, exec); err != nil {
		return err
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) introMsdos(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := IntroMsDos{}
	if err := m.Stat(ctx, exec); err != nil {
		return err
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) intro(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Intro{}
	if err := m.Stat(ctx, exec); err != nil {
		return err
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) installer(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Installer{}
	if err := m.Stat(ctx, exec); err != nil {
		return err
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) demoscene(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Demoscene{}
	if err := m.Stat(ctx, exec); err != nil {
		return err
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) nfo(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Nfo{}
	if err := m.Stat(ctx, exec); err != nil {
		return err
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) proof(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Proof{}
	if err := m.Stat(ctx, exec); err != nil {
		return err
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) ansi(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Ansi{}
	if err := m.Stat(ctx, exec); err != nil {
		return err
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) ansiBrand(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := AnsiBrand{}
	if err := m.Stat(ctx, exec); err != nil {
		return err
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) ansiBBS(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := AnsiBBS{}
	if err := m.Stat(ctx, exec); err != nil {
		return err
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) ansiFTP(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := AnsiFTP{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("ansiFTP.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) ansiPack(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := AnsiPack{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("ansiPack.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) ansiNfo(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := AnsiNfo{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("ansiNfo.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) bbs(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := BBS{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("bbs.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) bbstro(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := BBStro{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("bbstro.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) bbsImage(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := BBSImage{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("bbsImage.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) bbsText(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := BBSText{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("bbsText.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) ftp(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := FTP{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("ftp.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) magazine(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Magazine{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("magazine.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) text(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Text{}
	if err := m.Stat(ctx, exec); err != nil {
		return err
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) textPack(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := TextPack{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("textPack.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) imagePack(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := ImagePack{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("imagePack.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) windowsPack(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := WindowsPack{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("windowsPack.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) msdosPack(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := MsDosPack{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("msdosPack.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) database(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Database{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("database.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) textAmiga(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := TextAmiga{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("textAmiga.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) textApple2(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := TextApple2{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("textApple2.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) textAtariST(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := TextAtariST{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("textAtariST.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) pdf(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := PDF{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("pdf.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) html(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := HTML{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("html.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) newsArticle(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := NewsArticle{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("newsArticle.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) standards(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Standard{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("standards.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) announcement(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Announcement{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("announcement.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) jobAdvert(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := JobAdvert{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("jobAdvert.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) trialCrackme(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := TrialCrackme{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("trailCrackme.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) hack(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Hack{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("hack.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) tool(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Tool{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("tool.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) takedown(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Takedown{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("takedown.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) drama(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Drama{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("drama.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) advert(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Advert{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("advert.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) restrict(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Restrict{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("restrict.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) howTo(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := HowTo{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("howTo.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) nfoTool(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := NfoTool{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("nfoTool.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) image(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Image{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("image.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) music(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Music{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("music.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) video(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Video{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("video.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) msdos(ctx context.Context, exec boil.ContextExecutor) error {
	m := MsDos{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("msdos.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) windows(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Windows{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("windows.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) macos(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Macos{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("macos.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) linux(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Linux{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("linux.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) java(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Java{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("java.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}

func (s *Summary) script(ctx context.Context, exec boil.ContextExecutor) error {
	panics.BoilExecCrash(exec)
	m := Script{}
	if err := m.Stat(ctx, exec); err != nil {
		return fmt.Errorf("script.Stat: %w", err)
	}
	s.Update(m.Count, m.Bytes, m.MinYear, m.MaxYear)
	return nil
}
