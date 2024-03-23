package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/tags"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const (
	uidPlaceholder = `ADB7C2BF-7221-467B-B813-3636FE4AE16B` // UID of the user who deleted the file.
)

var (
	ErrRels     = errors.New("too many releasers, only two are allowed")
	ErrPlatform = errors.New("invalid platform")
	ErrTag      = errors.New("invalid tag")
	ErrYear     = errors.New("invalid year")
	ErrMonth    = errors.New("invalid month")
	ErrDay      = errors.New("invalid day")
)

// GetPlatformTagInfo returns the human readable platform and tag name.
func GetPlatformTagInfo(c echo.Context, platform, tag string) (string, error) {
	if c == nil {
		return "", ErrCtx
	}
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
func GetTagInfo(c echo.Context, tag string) (string, error) {
	if c == nil {
		return "", ErrCtx
	}
	t := tags.TagByURI(tag)
	if t == -1 {
		return "", fmt.Errorf("%s: %w", tag, ErrTag)
	}
	s := tags.Infos()[t]
	return s, nil
}

// UpdateOnline updates the record to be online and public.
func UpdateOnline(c echo.Context, id int64) error {
	if c == nil {
		return ErrCtx
	}
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return err
	}
	f.Deletedat = null.TimeFromPtr(nil)
	f.Deletedby = null.String{}
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return err
	}
	return nil
}

// UpdateOffline updates the record to be offline and inaccessible to the public.
func UpdateOffline(c echo.Context, id int64) error {
	if c == nil {
		return ErrCtx
	}
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return err
	}
	now := time.Now()
	f.Deletedat = null.TimeFromPtr(&now)
	f.Deletedby = null.StringFrom(strings.ToLower(uidPlaceholder))
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return err
	}
	return nil
}

// UpdateNoReadme updates the retrotxt_no_readme column value with val.
// It returns nil if the update was successful.
// Id is the database id of the record.
func UpdateNoReadme(c echo.Context, id int64, val bool) error {
	if c == nil {
		return ErrCtx
	}

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()

	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return err
	}

	i := int16(0)
	if val {
		i = 1
	}
	f.RetrotxtNoReadme = null.NewInt16(i, true)

	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return err
	}
	return nil
}

// UpdatePlatform updates the platform column value with val.
// It returns nil if the update was successful.
// Id is the database id of the record.
func UpdatePlatform(c echo.Context, id int64, val string) error {
	if c == nil {
		return ErrCtx
	}
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
		return err
	}
	f.Platform = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return err
	}
	return nil
}

// UpdateReleasers updates the releasers values with val.
// Two releases can be separated by a + (plus) character.
// It returns nil if the update was successful.
// Id is the database id of the record.
func UpdateReleasers(c echo.Context, id int64, val string) error {
	if c == nil {
		return ErrCtx
	}
	const max = 2
	val = strings.TrimSpace(val)
	s := strings.Split(val, "+")
	if len(s) > max {
		return fmt.Errorf("%s: %w", s, ErrRels)
	}

	for i, v := range s {
		s[i] = strings.ToUpper(releaser.Clean(v))
	}

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return err
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
func UpdateTag(c echo.Context, id int64, val string) error {
	if c == nil {
		return ErrCtx
	}

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
		return err
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
func UpdateTitle(c echo.Context, id int64, val string) error {
	if c == nil {
		return ErrCtx
	}
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := FindFile(ctx, db, id)
	if err != nil {
		return err
	}
	f.RecordTitle = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("%s: %w", val, err)
	}
	return nil
}

// UpdateYMD updates the title column value with val.
// It returns nil if the update was successful.
// Id is the database id of the record.
func UpdateYMD(c echo.Context, id int64, y, m, d null.Int16) error {
	if c == nil {
		return ErrCtx
	}

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
		return err
	}
	f.DateIssuedYear = y
	f.DateIssuedMonth = m
	f.DateIssuedDay = d
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return err
	}
	return nil
}
