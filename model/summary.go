//nolint:gochecknoglobals
package model

// Package file summary.go contains the database queries for the statistics of files.

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/Defacto2/server/handler/releaser/name"
	"github.com/Defacto2/server/internal/nils"
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
func (obj *Summary) ByDescription(ctx context.Context, exec boil.ContextExecutor, terms []string) error {
	const format = "summary by description: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	size := len(terms)
	args := make([]any, 0, size)
	orConditions := make([]string, 0, size)
	for i, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}

		args = append(args, term)
		orConditions = append(
			orConditions,
			"to_tsvector('english', concat_ws(' ', files.record_title, files.comment)) @@ "+
				"plainto_tsquery('english', $"+strconv.Itoa(i+1)+")",
		)
	}
	if len(orConditions) == 0 {
		return fmt.Errorf(format, ErrSearch)
	}

	// combine with proper parentheses for correct operator precedence
	query := string(postgres.Summary())
	query += "(" + strings.Join(orConditions, " OR ") + ") AND " + ClauseNoSoftDel
	query = strings.TrimSpace(query)

	return queries.Raw(query, args...).Bind(ctx, exec, obj)
}

// ByFilename saves the summary statistics for the filename search.
func (obj *Summary) ByFilename(ctx context.Context, exec boil.ContextExecutor, terms []string) error {
	const format = "summary by filename: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	const size = 4
	args := make([]any, 0, len(terms)*size)
	orConditions := make([]string, 0, len(terms))

	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}

		const (
			seq1 = iota + 1
			seq2
			seq3
			seq4
		)
		idx := len(args)
		// note using the sequence '~*' makes it non case-sensitive
		cond := "(filename ~* $" + strconv.Itoa(idx+seq1) +
			" OR filename ILIKE $" + strconv.Itoa(idx+seq2) +
			" OR filename ILIKE $" + strconv.Itoa(idx+seq3) +
			" OR filename ILIKE $" + strconv.Itoa(idx+seq4) + ")"
		orConditions = append(orConditions, cond)

		args = append(args, term, term+"%", "%"+term, "%"+term+"%")
	}

	if len(orConditions) == 0 {
		return fmt.Errorf(format, ErrSearch)
	}

	query := string(postgres.Summary()) + "(" + strings.Join(orConditions, " OR ") + ") AND " + ClauseNoSoftDel

	return queries.Raw(query, args...).Bind(ctx, exec, obj)
}

// ByForApproval returns the summary statistics for files that require approval.
func (obj *Summary) ByForApproval(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "summary by for approval: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	return models.NewQuery(
		models.FileWhere.Deletedat.IsNotNull(), models.FileWhere.Deletedby.IsNull(),
		qm.WithDeleted(),
		qm.Select(SummCols()...),
		qm.From(From),
	).Bind(ctx, exec, obj)
}

// ByHidden returns the summary statistics for files that have been deleted.
func (obj *Summary) ByHidden(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "summary by hidden: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	return models.NewQuery(
		models.FileWhere.Deletedat.IsNotNull(), models.FileWhere.Deletedby.IsNotNull(),
		qm.WithDeleted(),
		qm.Select(SummCols()...),
		qm.From(From),
	).Bind(ctx, exec, obj)
}

// ByPublic selects the summary statistics for all public files.
func (obj *Summary) ByPublic(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "summary by for approval: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	nils.BoilExecCrash(exec)
	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(ClauseNoSoftDel),
		qm.From(From),
	).Bind(ctx, exec, obj)
}

// ByScener selects the summary statistics for the named sceners.
func (obj *Summary) ByScener(ctx context.Context, exec boil.ContextExecutor, scener string) error {
	const format = "summary by scener: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	scener = strings.TrimSpace(scener)
	if scener == "" {
		return nil
	}
	query, params := postgres.ScenerSQL(scener)

	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where(query, params...),
		qm.Where(ClauseNoSoftDel),
		qm.From(From),
	).Bind(ctx, exec, obj)
}

// ByReleaser returns the summary statistics for the named releaser.
// The uri is case insensitive and must be the URI slug of the releaser
// or an error could be returned.
func (obj *Summary) ByReleaser(ctx context.Context, exec boil.ContextExecutor, uri string) error {
	const format = "summary by releaser: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	human, err := name.Humanize(name.Path(uri))
	if err != nil {
		return fmt.Errorf(format, err)
	}
	s := strings.ToUpper(human)
	arg := null.StringFrom(s)

	return models.NewQuery(
		qm.Select(SummCols()...),
		qm.Where("upper(group_brand_for) = ? OR upper(group_brand_by) = ?", arg, arg),
		qm.Where(ClauseNoSoftDel),
		qm.From(From),
	).Bind(ctx, exec, obj)
}

// ByUnwanted returns the summary statistics for files that have been marked as unwanted.
func (obj *Summary) ByUnwanted(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "summary by unwanted: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	empty := null.StringFrom("")

	return models.NewQuery(
		models.FileWhere.FileSecurityAlertURL.IsNotNull(),
		models.FileWhere.FileSecurityAlertURL.NEQ(empty),
		qm.WithDeleted(),
		qm.Select(SummCols()...),
		qm.From(From),
	).Bind(ctx, exec, obj)
}

// Update updates the summary statistics,
// however there are no sanity checks of the values.
func (obj *Summary) Update(count, bytes, yearMin, yearMax int) {
	abs16 := func(n int) int16 {
		if n < 0 {
			return int16(0)
		}
		if n > math.MaxInt16 {
			return math.MaxInt16
		}
		return int16(n)
	}

	obj.SumCount = sql.NullInt64{Int64: int64(count), Valid: true}
	obj.SumBytes = sql.NullInt64{Int64: int64(bytes), Valid: true}
	obj.MinYear = sql.NullInt16{Int16: abs16(yearMin), Valid: true}
	obj.MaxYear = sql.NullInt16{Int16: abs16(yearMax), Valid: true}
}

// ByMatch returns the summary statistics for the named uri.
func (obj *Summary) ByMatch(ctx context.Context, exec boil.ContextExecutor, uri string) error {
	if fn, ok := matches[Key(uri)]; ok {
		return fn(obj, ctx, exec)
	}

	const format = "summary by match %q: %w"
	return fmt.Errorf(format, uri, ErrURI)
}

// StatFunc is a function that updates the summary statistics.
type StatFunc func(*Summary, context.Context, boil.ContextExecutor) error

var matches = map[Key]StatFunc{
	KeyPCBoard:      (*Summary).pcboard,
	KeyPCBoardPPE:   (*Summary).pcboardPPE,
	KeyPCBoardText:  (*Summary).pcboardText,
	KeyTextAmiga:    (*Summary).textAmiga,
	KeyTextApple2:   (*Summary).textApple2,
	KeyTextAtariST:  (*Summary).textAtariST,
	KeyPDF:          (*Summary).pdf,
	KeyHTML:         (*Summary).html,
	KeyNewsArticle:  (*Summary).newsArticle,
	KeyStandards:    (*Summary).standards,
	KeyAnnouncement: (*Summary).announcement,
	KeyJobAdvert:    (*Summary).jobAdvert,
	KeyTrialCrackme: (*Summary).trialCrackme,
	KeyHack:         (*Summary).hack,
	KeyTool:         (*Summary).tool,
	KeyTakedown:     (*Summary).takedown,
	KeyDrama:        (*Summary).drama,
	KeyAdvert:       (*Summary).advert,
	KeyRestrict:     (*Summary).restrict,
	KeyHowTo:        (*Summary).howTo,
	KeyNFOTool:      (*Summary).nfoTool,
	KeyImage:        (*Summary).image,
	KeyMusic:        (*Summary).music,
	KeyVideo:        (*Summary).video,
	KeyMSDOS:        (*Summary).msdos,
	KeyWindows:      (*Summary).windows,
	KeyMacOS:        (*Summary).macos,
	KeyLinux:        (*Summary).linux,
	KeyJava:         (*Summary).java,
	KeyScript:       (*Summary).script,
	KeyDatabase:     (*Summary).database,
	KeyMSDOSPack:    (*Summary).msdosPack,
	KeyWindowsPack:  (*Summary).windowsPack,
	KeyImagePack:    (*Summary).imagePack,
	KeyTextPack:     (*Summary).textPack,
	KeyText:         (*Summary).text,
	KeyMagazine:     (*Summary).magazine,
	KeyFTP:          (*Summary).ftp,
	KeyBBSText:      (*Summary).bbsText,
	KeyBBSImage:     (*Summary).bbsImage,
	KeyBBStro:       (*Summary).bbstro,
	KeyBBS:          (*Summary).bbs,
	KeyANSINFO:      (*Summary).ansiNfo,
	KeyANSIPack:     (*Summary).ansiPack,
	KeyANSIFTP:      (*Summary).ansiFTP,
	KeyANSIBBS:      (*Summary).ansiBBS,
	KeyANSIBrand:    (*Summary).ansiBrand,
	KeyANSI:         (*Summary).ansi,
	KeyProof:        (*Summary).proof,
	KeyNFO:          (*Summary).nfo,
	KeyDemoscene:    (*Summary).demoscene,
	KeyInstaller:    (*Summary).installer,
	KeyIntro:        (*Summary).intro,
	KeyIntroMSDOS:   (*Summary).introMsdos,
	KeyIntroWindows: (*Summary).introWindows,
	KeyConsole:      (*Summary).console,
}

// Keys is intended for testing boilerplate and returns the keys used in matches.
func Keys() []Key {
	return slices.Collect(maps.Keys(matches))
}

type StatModel interface {
	Stat(ctx context.Context, exec boil.ContextExecutor) error
	Values() (count, bytes, minYear, maxYear int)
}

func execStat[T any, PT interface {
	*T
	StatModel
}](ctx context.Context, exec boil.ContextExecutor, obj *Summary, key Key) error {
	const format = "%s: %w"
	if err := nils.Check(ctx, exec, obj); err != nil {
		return fmt.Errorf(format, string(key), err)
	}

	var m T
	filter := PT(&m)

	if err := filter.Stat(ctx, exec); err != nil {
		return fmt.Errorf(format, string(key), err)
	}
	obj.Update(filter.Values())
	return nil
}

func (obj *Summary) introWindows(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[IntroWindows](ctx, exec, obj, KeyIntroWindows)
}

func (obj *Summary) script(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Script](ctx, exec, obj, KeyScript)
}

func (obj *Summary) console(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Console](ctx, exec, obj, KeyConsole)
}

func (obj *Summary) pcboard(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[PCBoard](ctx, exec, obj, KeyPCBoard)
}

func (obj *Summary) pcboardPPE(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[PCBoardPPE](ctx, exec, obj, KeyPCBoardPPE)
}

func (obj *Summary) pcboardText(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[PCBoardText](ctx, exec, obj, KeyPCBoardText)
}

func (obj *Summary) introMsdos(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[IntroMsDos](ctx, exec, obj, KeyIntroMSDOS)
}

func (obj *Summary) intro(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Intro](ctx, exec, obj, KeyIntro)
}

func (obj *Summary) installer(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Installer](ctx, exec, obj, KeyInstaller)
}

func (obj *Summary) demoscene(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Demoscene](ctx, exec, obj, KeyDemoscene)
}

func (obj *Summary) nfo(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Nfo](ctx, exec, obj, KeyNFO)
}

func (obj *Summary) proof(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Proof](ctx, exec, obj, KeyProof)
}

func (obj *Summary) ansi(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Ansi](ctx, exec, obj, KeyANSI)
}

func (obj *Summary) ansiBrand(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[AnsiBrand](ctx, exec, obj, KeyANSIBrand)
}

func (obj *Summary) ansiBBS(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[AnsiBBS](ctx, exec, obj, KeyANSIBBS)
}

func (obj *Summary) ansiFTP(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[AnsiFTP](ctx, exec, obj, KeyANSIFTP)
}

func (obj *Summary) ansiPack(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[AnsiPack](ctx, exec, obj, KeyANSIPack)
}

func (obj *Summary) ansiNfo(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[AnsiNfo](ctx, exec, obj, KeyANSINFO)
}

func (obj *Summary) bbs(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[BBS](ctx, exec, obj, KeyBBS)
}

func (obj *Summary) bbstro(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[BBStro](ctx, exec, obj, KeyBBStro)
}

func (obj *Summary) bbsImage(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[BBSImage](ctx, exec, obj, KeyBBSImage)
}

func (obj *Summary) bbsText(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[BBSText](ctx, exec, obj, KeyBBSText)
}

func (obj *Summary) ftp(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[FTP](ctx, exec, obj, KeyFTP)
}

func (obj *Summary) magazine(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Magazine](ctx, exec, obj, KeyMagazine)
}

func (obj *Summary) text(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Text](ctx, exec, obj, KeyText)
}

func (obj *Summary) textPack(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[TextPack](ctx, exec, obj, KeyTextPack)
}

func (obj *Summary) imagePack(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[ImagePack](ctx, exec, obj, KeyImagePack)
}

func (obj *Summary) windowsPack(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[WindowsPack](ctx, exec, obj, KeyWindowsPack)
}

func (obj *Summary) msdosPack(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[MsDosPack](ctx, exec, obj, KeyMSDOSPack)
}

func (obj *Summary) database(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Database](ctx, exec, obj, KeyDatabase)
}

func (obj *Summary) textAmiga(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[TextAmiga](ctx, exec, obj, KeyTextAmiga)
}

func (obj *Summary) textApple2(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[TextApple2](ctx, exec, obj, KeyTextApple2)
}

func (obj *Summary) textAtariST(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[TextAtariST](ctx, exec, obj, KeyTextAtariST)
}

func (obj *Summary) pdf(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[PDF](ctx, exec, obj, KeyPDF)
}

func (obj *Summary) html(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[HTML](ctx, exec, obj, KeyHTML)
}

func (obj *Summary) newsArticle(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[NewsArticle](ctx, exec, obj, KeyNewsArticle)
}

func (obj *Summary) standards(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Standards](ctx, exec, obj, KeyStandards)
}

func (obj *Summary) announcement(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Announcement](ctx, exec, obj, KeyAnnouncement)
}

func (obj *Summary) jobAdvert(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[JobAdvert](ctx, exec, obj, KeyJobAdvert)
}

func (obj *Summary) trialCrackme(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[TrialCrackme](ctx, exec, obj, KeyTrialCrackme)
}

func (obj *Summary) hack(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Hack](ctx, exec, obj, KeyHack)
}

func (obj *Summary) tool(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Tool](ctx, exec, obj, KeyTool)
}

func (obj *Summary) takedown(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Takedown](ctx, exec, obj, KeyTakedown)
}

func (obj *Summary) drama(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Drama](ctx, exec, obj, KeyDrama)
}

func (obj *Summary) advert(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Advert](ctx, exec, obj, KeyAdvert)
}

func (obj *Summary) restrict(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Restrict](ctx, exec, obj, KeyRestrict)
}

func (obj *Summary) howTo(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[HowTo](ctx, exec, obj, KeyHowTo)
}

func (obj *Summary) nfoTool(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[NfoTool](ctx, exec, obj, KeyNFOTool)
}

func (obj *Summary) image(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Image](ctx, exec, obj, KeyImage)
}

func (obj *Summary) music(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Music](ctx, exec, obj, KeyMusic)
}

func (obj *Summary) video(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Video](ctx, exec, obj, KeyVideo)
}

func (obj *Summary) msdos(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[MsDos](ctx, exec, obj, KeyMSDOS)
}

func (obj *Summary) windows(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Windows](ctx, exec, obj, KeyWindows)
}

func (obj *Summary) macos(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Macos](ctx, exec, obj, KeyMacOS)
}

func (obj *Summary) linux(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Linux](ctx, exec, obj, KeyLinux)
}

func (obj *Summary) java(ctx context.Context, exec boil.ContextExecutor) error {
	return execStat[Java](ctx, exec, obj, KeyJava)
}
