package model

import (
	"context"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const (
	uidPlaceholder = `ADB7C2BF-7221-467B-B813-3636FE4AE16B` // UID of the user who deleted the file.
)

// GetPlatformTagInfo returns the human readable platform and tag name.
func GetPlatformTagInfo(c echo.Context, platform, tag string) (string, error) {
	if c == nil {
		return "", ErrCtx
	}
	p, t := tags.TagByURI(platform), tags.TagByURI(tag)
	return tags.Humanize(p, t), nil
}

// GetTagInfo returns the human readable tag name.
func GetTagInfo(c echo.Context, tag string) (string, error) {
	if c == nil {
		return "", ErrCtx
	}
	t := tags.TagByURI(tag)
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
	f, err := models.FindFile(ctx, db, id)
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
	f, err := models.FindFile(ctx, db, id)
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
	f, err := models.FindFile(ctx, db, id)
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

	// TODO: validate val against a list of platforms
	val = strings.ToLower(val)

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := models.FindFile(ctx, db, id)
	if err != nil {
		return err
	}
	f.Platform = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return err
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

	// TODO: validate val against a list of SECTIONS
	val = strings.ToLower(val)

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := models.FindFile(ctx, db, id)
	if err != nil {
		return err
	}
	f.Section = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return err
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
	f, err := models.FindFile(ctx, db, id)
	if err != nil {
		return err
	}
	// TODO: format val text
	f.RecordTitle = null.StringFrom(val)
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return err
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
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := models.FindFile(ctx, db, id)
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
