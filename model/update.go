package model

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/tags"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// GetPlatformTagInfo returns the human readable platform and tag name.
func GetPlatformTagInfo(platform, tag string) (string, error) {
	p, t := tags.TagByURI(platform), tags.TagByURI(tag)
	if p == -1 {
		return "", fmt.Errorf("%s: %w", platform, ErrPlatform)
	}
	if t == -1 {
		return "", fmt.Errorf("%s: %w", tag, ErrTag)
	}
	return tags.Humanize(p, t), nil
}

// GetTagInfo returns the human readable tag name.
func GetTagInfo(tag string) (string, error) {
	t := tags.TagByURI(tag)
	if t == -1 {
		return "", fmt.Errorf("%s: %w", tag, ErrTag)
	}
	s := tags.Infos()[t]
	return s, nil
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
		return fmt.Errorf("%s: %w", tag, ErrTag)
	}
	if !tags.IsTag(tag) {
		return fmt.Errorf("%s: %w", tag, ErrTag)
	}

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.Platform = null.StringFrom(p.String())
	f.Section = null.StringFrom(t.String())
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("f.update: %w", err)
	}
	return nil
}

func UpdateDateIssued(id int64, y, m, d string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	// todo: confirm year is valid

	year, month, day := DateIssue(y, m, d)
	f.DateIssuedYear = year
	f.DateIssuedMonth = month
	f.DateIssuedDay = day
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%q %q %q: %w", y, m, d, err)
	}
	return nil
}

func UpdateCreatorText(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.CreditText = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

func UpdateCreatorIll(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.CreditIllustration = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

func UpdateCreatorProg(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.CreditProgram = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

func UpdateCreatorAudio(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.CreditAudio = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

func UpdateCreators(id int64, text, ill, prog, audio string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.CreditText = null.StringFrom(text)
	f.CreditIllustration = null.StringFrom(ill)
	f.CreditProgram = null.StringFrom(prog)
	f.CreditAudio = null.StringFrom(audio)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", "updatecreators", err)
	}
	return nil
}

func UpdateComment(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	val = strings.TrimSpace(val)
	f.Comment = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

func UpdateYouTube(id int64, watch string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	// todo move to form, confirm length and chars
	watch = strings.TrimSpace(watch)
	f.WebIDYoutube = null.StringFrom(watch)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", watch, err)
	}
	return nil
}

func UpdateDemozoo(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	// todo validate range
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	f.WebIDDemozoo = null.Int64From(i)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

func UpdatePouet(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	// todo validate range
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	f.WebIDPouet = null.Int64From(i)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

func Update16Colors(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	// todo validate 16colors, drop url etc
	val = strings.TrimSpace(val)
	f.WebID16colors = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

func UpdateGitHub(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	// todo validate github repo name/repo fmt etc
	val = strings.TrimSpace(val)
	f.WebIDGithub = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

func UpdateRelations(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	// todo validate github repo name/repo fmt etc
	val = strings.TrimSpace(val)
	f.ListRelations = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

func UpdateSites(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	// todo validate github repo name/repo fmt etc
	val = strings.TrimSpace(val)
	f.ListLinks = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

// UpdateFilename updates the filename column value with val.
// It returns nil if the update was successful.
// Id is the database id of the record.
func UpdateFilename(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.Filename = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

// UpdateOnline updates the record to be online and public.
func UpdateOnline(id int64) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.Deletedat = null.TimeFromPtr(nil)
	f.Deletedby = null.String{}
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("f.update: %w", err)
	}
	return nil
}

// UpdateOffline updates the record to be offline and inaccessible to the public.
func UpdateOffline(id int64) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	now := time.Now()
	f.Deletedat = null.TimeFromPtr(&now)
	f.Deletedby = null.StringFrom(strings.ToLower(uidPlaceholder))
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("f.update: %w", err)
	}
	return nil
}

// UpdateNoReadme updates the retrotxt_no_readme column value with val.
// It returns nil if the update was successful.
// Id is the database id of the record.
func UpdateNoReadme(id int64, val bool) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()

	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}

	i := int16(0)
	if val {
		i = 1
	}
	f.RetrotxtNoReadme = null.NewInt16(i, true)

	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("f.update: %w", err)
	}
	return nil
}

// UpdatePlatform updates the platform column value with val.
// It returns nil if the update was successful.
// Id is the database id of the record.
func UpdatePlatform(id int64, val string) error {
	val = strings.ToLower(val)
	if p := tags.TagByURI(val); p == -1 {
		return fmt.Errorf("%s: %w", val, ErrPlatform)
	}

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.Platform = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("f.update: %w", err)
	}
	return nil
}

// UpdateReleasers updates the releasers values with val.
// Two releases can be separated by a + (plus) character.
// It returns nil if the update was successful.
// Id is the database id of the record.
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

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
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
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

// UpdateTag updates the section column value with val.
// It returns nil if the update was successful.
// Id is the database id of the record.
func UpdateTag(id int64, val string) error {
	val = strings.ToLower(val)
	if t := tags.TagByURI(val); t == -1 {
		return fmt.Errorf("%s: %w", val, ErrTag)
	}

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.Section = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

// UpdateTitle updates the title column value with val.
// It returns nil if the update was successful.
// Id is the database id of the record.
func UpdateTitle(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.RecordTitle = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

func UpdateVirusTotal(id int64, val string) error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	val = strings.TrimSpace(val)
	const link = "https://www.virustotal.com/"
	if len(val) > 0 && !strings.HasPrefix(val, link) {
		var ErrURL = fmt.Errorf("value must start with " + link)
		return fmt.Errorf("%s: %w", val, ErrURL)
	}
	f.FileSecurityAlertURL = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

// UpdateYMD updates the title column value with val.
// It returns nil if the update was successful.
// Id is the database id of the record.
func UpdateYMD(id int64, y, m, d null.Int16) error {
	if !y.IsZero() && !helper.IsYear(int(y.Int16)) {
		return fmt.Errorf("%d: %w", y.Int16, ErrYear)
	}
	if !m.IsZero() && helper.ShortMonth(int(m.Int16)) == "" {
		return fmt.Errorf("%d: %w", m.Int16, ErrMonth)
	}
	if !d.IsZero() && !helper.IsDay(int(d.Int16)) {
		return fmt.Errorf("%d: %w", d.Int16, ErrDay)
	}

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	f.DateIssuedYear = y
	f.DateIssuedMonth = m
	f.DateIssuedDay = d
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("f.update: %w", err)
	}
	return nil
}
