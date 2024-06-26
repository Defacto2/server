package model

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/demozoo"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/pouet"
	"github.com/Defacto2/server/internal/tags"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// int64From is a type for the int64 columns that can be updated.
type int64From int

const (
	demozooProd int64From = iota
	pouetProd
)

// Update16Colors updates the WebID16colors column value with val.
func Update16Colors(id int64, val string) error {
	return UpdateStringFrom(colors16, id, val)
}

// UpdateComment updates the Comment column value with val.
func UpdateComment(id int64, val string) error {
	return UpdateStringFrom(comment, id, val)
}

// UpdateCreatorAudio updates the CreditAudio column with val.
func UpdateCreatorAudio(id int64, val string) error {
	return UpdateStringFrom(credAudio, id, val)
}

// UpdateCreatorIll updates the CreditIllustration column with val.
func UpdateCreatorIll(id int64, val string) error {
	return UpdateStringFrom(credIll, id, val)
}

// UpdateCreatorProg updates the CreditProgram column with val.
func UpdateCreatorProg(id int64, val string) error {
	return UpdateStringFrom(credProg, id, val)
}

// UpdateCreatorText updates the CreditText column with val.
func UpdateCreatorText(id int64, val string) error {
	return UpdateStringFrom(creText, id, val)
}

// UpdateDemozoo updates the WebIDDemozoo column with val.
func UpdateDemozoo(id int64, val string) error {
	return UpdateInt64From(demozooProd, id, val)
}

// UpdateFilename updates the Filename column with val.
func UpdateFilename(id int64, val string) error {
	return UpdateStringFrom(filename, id, val)
}

// UpdateGitHub updates the WebIDGithub column with val.
func UpdateGitHub(id int64, val string) error {
	return UpdateStringFrom(github, id, val)
}

// UpdatePlatform updates the Platform column value with val.
func UpdatePlatform(id int64, val string) error {
	return UpdateStringFrom(platform, id, val)
}

// UpdatePouet updates the WebIDPouet column with val.
func UpdatePouet(id int64, val string) error {
	return UpdateInt64From(pouetProd, id, val)
}

// UpdateRelations updates the ListRelations column value with val.
func UpdateRelations(id int64, val string) error {
	return UpdateStringFrom(relations, id, val)
}

// UpdateSites updates the ListLinks column with val.
func UpdateSites(id int64, val string) error {
	return UpdateStringFrom(sites, id, val)
}

// UpdateTag updates the Section column with val.
func UpdateTag(id int64, val string) error {
	return UpdateStringFrom(section, id, val)
}

// UpdateTitle updates the RecordTitle column with val.
func UpdateTitle(id int64, val string) error {
	return UpdateStringFrom(title, id, val)
}

// UpdateVirusTotal updates the FileSecurityAlertURL value with val.
func UpdateVirusTotal(id int64, val string) error {
	return UpdateStringFrom(virusTotal, id, val)
}

// UpdateYouTube updates the WebIDYoutube column value with val.
func UpdateYouTube(id int64, val string) error {
	return UpdateStringFrom(youtube, id, val)
}

// UpdateInt64From updates the column int64 from value with val.
// The int64From columns are table columns that can either be null, empty, or have an int64 value.
// The demoZooProd and pouetProd values are also validated to be within a sane range.
func UpdateInt64From(column int64From, id int64, val string) error {
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return fmt.Errorf("updateint64from: %w", ErrDB)
	}
	defer db.Close()
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("find file for %q: %w", column, err)
	}

	i64, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}

	var invalid bool
	switch column {
	case demozooProd:
		invalid = i64 < 0 || i64 > demozoo.Sanity
		f.WebIDDemozoo = null.Int64From(i64)
	case pouetProd:
		invalid = i64 < 0 || i64 > pouet.Sanity
		f.WebIDPouet = null.Int64From(i64)
	default:
		return fmt.Errorf("updateint64from: %w", ErrColumn)
	}
	if invalid {
		return fmt.Errorf("%d: %w", i64, ErrID)
	}
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%q %s: %w", column, val, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("updateint64from: %w", err)
	}
	return nil
}

// stringFrom is a type for the string columns that can be updated.
type stringFrom int

const (
	colors16 stringFrom = iota
	comment
	credAudio
	credIll
	credProg
	creText
	filename
	github
	platform
	relations
	section
	sites
	title
	virusTotal
	youtube
)

// UpdateStringFrom updates the column string from value with val.
// The stringFrom columns are table columns that can either be null, empty, or have a string value.
func UpdateStringFrom(column stringFrom, id int64, val string) error {
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return fmt.Errorf("updatestringfrom: %w", ErrDB)
	}
	defer db.Close()
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("find file for %q: %w", column, err)
	}
	if err = updateStringCases(f, column, val); err != nil {
		return fmt.Errorf("updatestringfrom: %w", err)
	}
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%q %s: %w", column, val, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("updatestringfrom: %w", err)
	}
	return nil
}

func updateStringCases(f *models.File, column stringFrom, val string) error {
	s := null.StringFrom(strings.TrimSpace(val))
	switch column {
	case colors16:
		f.WebID16colors = s
	case comment:
		f.Comment = s
	case credAudio:
		f.CreditAudio = s
	case credIll:
		f.CreditIllustration = s
	case credProg:
		f.CreditProgram = s
	case creText:
		f.CreditText = s
	case filename:
		f.Filename = s
	case github:
		f.WebIDGithub = s
	case platform:
		f.Platform = s
	case relations:
		f.ListRelations = s
	case section:
		f.Section = s
	case sites:
		f.ListLinks = s
	case title:
		f.RecordTitle = s
	case virusTotal:
		f.FileSecurityAlertURL = s
	case youtube:
		f.WebIDYoutube = s
	default:
		return ErrColumn
	}
	return nil
}

// UpdateCreators updates the text, illustration, program, and audio credit columns with the values provided.
func UpdateCreators(id int64, text, ill, prog, audio string) error {
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return fmt.Errorf("updatecreators: %w", ErrDB)
	}
	defer db.Close()
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("updatecreators find file, %d: %w", id, err)
	}
	f.CreditText = null.StringFrom(text)
	f.CreditIllustration = null.StringFrom(ill)
	f.CreditProgram = null.StringFrom(prog)
	f.CreditAudio = null.StringFrom(audio)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", "updatecreators", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("updatecreators: %w", err)
	}
	return nil
}

// UpdateLinks updates the youtube, 16colors, relations, sites, demozoo, and pouet columns with the values provided.
func UpdateLinks(id int64, youtube, colors16, github, relations, sites string, demozoo, pouet int64) error {
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return fmt.Errorf("updatelinks: %w", ErrDB)
	}
	defer db.Close()
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("updatelinks find file, %d: %w", id, err)
	}
	f.WebIDYoutube = null.StringFrom(youtube)
	f.WebID16colors = null.StringFrom(colors16)
	f.WebIDGithub = null.StringFrom(github)
	f.ListRelations = null.StringFrom(relations)
	f.ListLinks = null.StringFrom(sites)
	f.WebIDDemozoo = null.Int64From(demozoo)
	f.WebIDPouet = null.Int64From(pouet)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", "updatelinks", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("updatelinks: %w", err)
	}
	return nil
}

// UpdateClassification updates the classification of a file in the database.
// It takes an ID, platform, and tag as parameters and returns an error if any.
// Both platform and tag must be valid values.
func UpdateClassification(id int64, platform, tag string) error {
	p, t := tags.TagByURI(platform), tags.TagByURI(tag)
	if p == -1 {
		return fmt.Errorf("%s: %w", platform, ErrPlatform)
	}
	if !tags.IsPlatform(platform) {
		return fmt.Errorf("%s: %w", platform, ErrPlatform)
	}
	if t == -1 {
		return fmt.Errorf("%s: %w", tag, tags.ErrTag)
	}
	if !tags.IsTag(tag) {
		return fmt.Errorf("%s: %w", tag, tags.ErrTag)
	}
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.Platform = null.StringFrom(p.String())
	f.Section = null.StringFrom(t.String())
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("f.update: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("tx.commit: %w", err)
	}
	return nil
}

// UpdateDateIssued updates the date issued year, month and day columns with the values provided.
// Columns updated are DateIssuedYear, DateIssuedMonth, and DateIssuedDay.
func UpdateDateIssued(id int64, y, m, d string) error {
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return fmt.Errorf("update date issued: %w", err)
	}
	defer db.Close()
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	year, month, day := ValidDateIssue(y, m, d)
	f.DateIssuedYear = year
	f.DateIssuedMonth = month
	f.DateIssuedDay = day
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%q %q %q: %w", y, m, d, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("tx.commit: %w", err)
	}
	return nil
}

// UpdateNoReadme updates the retrotxt_no_readme column value with val.
// It returns nil if the update was successful.
// Id is the database id of the record.
func UpdateNoReadme(id int64, val bool) error {
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return fmt.Errorf("update noreadme: %w", err)
	}
	defer db.Close()
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	i := int16(0)
	if val {
		i = 1
	}
	f.RetrotxtNoReadme = null.NewInt16(i, true)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("f.update: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("tx.commit: %w", err)
	}
	return nil
}

// UpdateOffline updates the record to be offline and inaccessible to the public.
func UpdateOffline(id int64) error {
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return fmt.Errorf("update offline: %w", err)
	}
	defer db.Close()
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	now := time.Now()
	f.Deletedat = null.TimeFromPtr(&now)
	f.Deletedby = null.StringFrom(strings.ToLower(uidPlaceholder))
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("f.update: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("tx.commit: %w", err)
	}
	return nil
}

// UpdateOnline updates the record to be online and public.
func UpdateOnline(id int64) error {
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return fmt.Errorf("update online: %w", err)
	}
	defer db.Close()
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.Deletedat = null.TimeFromPtr(nil)
	f.Deletedby = null.String{}
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("f.update: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("tx.commit: %w", err)
	}
	return nil
}

// UpdateReleasers updates the releasers values with val.
// Two releases can be separated by a + (plus) character.
// The columns updated are GroupBrandFor and GroupBrandBy.
func UpdateReleasers(id int64, val string) error {
	const max = 2
	val = strings.TrimSpace(val)
	s := strings.Split(val, "+")
	if len(s) > max {
		return fmt.Errorf("%s: %w", s, ErrRels)
	}
	for i, v := range s {
		s[i] = releaser.Cell(v)
	}
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return fmt.Errorf("update releasers: %w", err)
	}
	defer db.Close()
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	switch len(s) {
	case max:
		f.GroupBrandFor = null.StringFrom(s[0])
		f.GroupBrandBy = null.StringFrom(s[1])
	case 1:
		f.GroupBrandFor = null.StringFrom(s[0])
		f.GroupBrandBy = null.StringFrom("")
	case 0:
		f.GroupBrandFor = null.StringFrom("")
		f.GroupBrandBy = null.StringFrom("")
	}
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("tx.commit: %w", err)
	}
	return nil
}

// UpdateYMD updates the date issued year, month and day columns with the values provided.
func UpdateYMD(ctx context.Context, exec boil.ContextExecutor, id int64, y, m, d null.Int16) error {
	if id <= 0 {
		return fmt.Errorf("updateymd id value %w: %d", ErrKey, id)
	}
	if !y.IsZero() && !helper.Year(int(y.Int16)) {
		return fmt.Errorf("updateymd %w: %d", ErrYear, y.Int16)
	}
	if !m.IsZero() && helper.ShortMonth(int(m.Int16)) == "" {
		return fmt.Errorf("updateymd %w: %d", ErrMonth, m.Int16)
	}
	if !d.IsZero() && !helper.Day(int(d.Int16)) {
		return fmt.Errorf("updateymd %w: %d", ErrDay, d.Int16)
	}
	f, err := OneFile(ctx, exec, id)
	if err != nil {
		return fmt.Errorf("updateymd one file %w: %d", err, id)
	}
	f.DateIssuedYear = y
	f.DateIssuedMonth = m
	f.DateIssuedDay = d
	if _, err = f.Update(ctx, exec, boil.Infer()); err != nil {
		return fmt.Errorf("updateymd update %w: %d", err, id)
	}
	return nil
}
