package model

// Package file inserts.go contains the database queries for inserting new file records.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// InsertDemozoo inserts a new file record into the database using a Demozoo production ID.
// This will not check if the Demozoo production ID already exists in the database.
// When successful the function will return the new record ID.
func InsertDemozoo(ctx context.Context, exec boil.ContextExecutor, prodID int) (key int64, unid string, err error) {
	const format = "insert by demozoo %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return 0, "", fmt.Errorf(format, "check", err)
	}
	if prodID < startID || prodID > demozoo.Sanity {
		return 0, "", fmt.Errorf(format, strconv.Itoa(prodID), ErrBadID)
	}

	now, newUUID, err := NewV7()
	if err != nil {
		return 0, "", fmt.Errorf(format, "uuid v7", err)
	}
	unid = newUUID.String()

	//nolint:exhaustruct // Only setting essential fields for database insertion
	f := models.File{
		UUID:         null.StringFrom(unid),
		WebIDDemozoo: null.Int64From(int64(math.Abs(float64(prodID)))),
		Deletedat:    null.TimeFromPtr(&now),
	}

	if err = f.Insert(ctx, exec, boil.Infer()); err != nil {
		return 0, "", fmt.Errorf(format, "infer", err)
	}

	key = f.ID

	return key, unid, nil
}

// InsertPouet inserts a new file record into the database using a Pouet production ID.
// This will not check if the Pouet production ID already exists in the database.
// When successful the function will return the new record ID.
func InsertPouet(ctx context.Context, exec boil.ContextExecutor, prodID int) (key int64, unid string, err error) {
	const format = "insert by pouet %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return 0, "", fmt.Errorf(format, "check", err)
	}
	if prodID < startID || prodID > pouet.Sanity {
		return 0, "", fmt.Errorf(format, strconv.Itoa(prodID), ErrBadID)
	}

	now, newUUID, err := NewV7()
	if err != nil {
		return 0, "", fmt.Errorf(format, "uuid v7", err)
	}
	unid = newUUID.String()

	//nolint:exhaustruct // Only setting essential fields for database insertion
	f := models.File{
		UUID:       null.StringFrom(unid),
		WebIDPouet: null.Int64From(int64(math.Abs(float64(prodID)))),
		Deletedat:  null.TimeFromPtr(&now),
	}

	if err = f.Insert(ctx, exec, boil.Infer()); err != nil {
		return 0, "", fmt.Errorf(format, "infer", err)
	}

	key = f.ID

	return key, unid, nil
}

// InsertUpload inserts a new file record into the database using a URL values map.
// This will not check if the file already exists in the database.
// Invalid values will be ignored, but will not prevent the record from being inserted.
// When successful the function will return the new record ID key and the UUID.
func InsertUpload(ctx context.Context, exec boil.ContextExecutor, values url.Values, prefix string) (
	key int64, unid uuid.UUID, err error,
) {
	none := uuid.UUID{}
	const format = "insert upload %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return 0, none, fmt.Errorf(format, "check", err)
	}

	now, unid, err := NewV7()
	if err != nil {
		return 0, none, fmt.Errorf(format, "uuid v7", err)
	}
	uid := unid.String()

	unique := null.StringFrom(uid)
	if exist, err := ExistUUID(ctx, exec, uid); err != nil {
		return 0, none, fmt.Errorf(format, "uuid exists", err)
	} else if exist {
		return 0, none, fmt.Errorf(format, uid, ErrUUID)
	}

	deleteT := null.TimeFromPtr(&now)
	if !deleteT.Valid || deleteT.Time.IsZero() {
		return 0, none, fmt.Errorf(format+" %v", format, ErrTime, deleteT.Time)
	}
	createT := null.TimeFromPtr(&now)
	if !createT.Valid || createT.Time.IsZero() {
		return 0, none, fmt.Errorf(format+" %v", format, ErrTime, createT.Time)
	}

	//nolint:exhaustruct // Only setting essential fields for database insertion
	f := models.File{
		UUID:      unique,
		Deletedat: deleteT,
		Createdat: createT,
	}

	f, err = validateUpload(f, values, prefix)
	if err != nil {
		return 0, none, fmt.Errorf(format, "validation", err)
	}
	if err = f.Insert(ctx, exec, boil.Infer()); err != nil {
		return 0, none, fmt.Errorf(format, prefix, err)
	}

	key = f.ID

	return key, unid, nil
}

func validateUpload(f models.File, values url.Values, prefix string) (models.File, error) {
	const format = "upload validate %s: %w"

	if v := values.Get(prefix + "-youtube"); v != "" {
		youtube := ValidYouTube(v)
		if !youtube.Valid {
			return f, fmt.Errorf(format, "youtube", ErrBadYT)
		}
		f.WebIDYoutube = youtube
	}

	releaser1, releaser2 := ValidReleasers(
		values.Get(prefix+"-releaser1"),
		values.Get(prefix+"-releaser2"),
	)

	title := ValidTitle(values.Get(prefix + "-title"))

	year, month, day := ValidDateIssue(
		values.Get(prefix+"-year"),
		values.Get(prefix+"-month"),
		values.Get(prefix+"-day"),
	)

	fname := values.Get(prefix + "-filename")
	filename := ValidFilename(fname)
	if !filename.Valid || filename.IsZero() {
		return f, fmt.Errorf(format, prefix+"-filename is required", ErrName)
	}

	filesize, err := ValidFilesize(values.Get(prefix + "-size"))
	if err != nil {
		return f, fmt.Errorf(format, "file size", err)
	}

	f.GroupBrandFor = releaser1
	f.GroupBrandBy = releaser2
	f.RecordTitle = title
	f.DateIssuedYear = year
	f.DateIssuedMonth = month
	f.DateIssuedDay = day
	f.Filename = filename
	f.Filesize = filesize

	readme := ValidFilename(values.Get(prefix + "-readme"))
	f.RetrotxtReadme = readme
	filemagic := ValidMagic(values.Get(prefix + "-magic"))
	f.FileMagicType = filemagic
	integrity := ValidIntegrity(values.Get(prefix + "-integrity"))
	f.FileIntegrityStrong = integrity

	lastMod := ValidLastMod(values.Get(prefix + "-lastmodified"))
	f.FileLastModified = lastMod
	content := ValidString(values.Get(prefix + "-content"))
	f.FileZipContent = fixZIPEntries(content)

	platform := ValidPlatform(values.Get(prefix + "-operating-system"))
	section := ValidSection(values.Get(prefix + "-category"))
	f.Platform = platform
	f.Section = SiteAd(releaser1, section)

	creditT := ValidSceners(values.Get(prefix + "-credittext"))
	f.CreditText = creditT
	creditI := ValidSceners(values.Get(prefix + "-creditill"))
	f.CreditIllustration = creditI
	creditP := ValidSceners(values.Get(prefix + "-creditprog"))
	f.CreditProgram = creditP
	creditA := ValidSceners(values.Get(prefix + "-creditaudio"))
	f.CreditAudio = creditA

	return f, nil
}

// SiteAd will replace a tags.Nfo section to either tags.BBS or tags.Ftp if the releaser
// is a known BBS board or FTP site. Otherwise the supplied section is returned.
func SiteAd(releaser, section null.String) null.String {
	if !strings.EqualFold(section.String, tags.Nfo.String()) {
		return section
	}

	rel := strings.TrimSpace(strings.ToLower(releaser.String))
	if strings.HasSuffix(rel, " ftp") {
		return null.StringFrom(tags.Ftp.String())
	}

	if strings.HasSuffix(rel, " bbs") {
		return null.StringFrom(tags.BBS.String())
	}

	return section
}

// fixZIPEntries fixes the file content for ZIP files that have DOS file or directory names
// encoded in CP437 or Windows-1252, that sometimes have non-existent runes.
func fixZIPEntries(content null.String) null.String {
	if !content.Valid {
		return content
	}

	p := []byte(content.String)
	s, err := decodeEntries(p)
	if err != nil {
		return null.StringFrom("")
	}

	return null.StringFrom(string(s))
}

// decode entries replaces any problematic code page characters with a "?".
func decodeEntries(p []byte) ([]byte, error) {
	const format = "decode entries: %w"
	if utf8.Valid(p) {
		return p, nil
	}

	// decoder := charmap.CodePage437.NewDecoder()
	decoder := charmap.Windows1252.NewDecoder()
	s, err := io.ReadAll(transform.NewReader(bytes.NewReader(p), decoder))
	if err != nil {
		return nil, fmt.Errorf(format, err)
	}

	s = bytes.ToValidUTF8(s, []byte("?"))

	return s, nil
}

// NewV7 generates a new UUID version 7,
// if it fails then it fallbacks to a UUID version 1.
//
// It also returns the current time.
func NewV7() (time.Time, uuid.UUID, error) {
	now := time.Now()
	uid, err := uuid.NewV7()
	if err == nil {
		return now, uid, nil
	}

	const format = "new uuid v7 %w: %w"
	uid, err = uuid.NewUUID()
	if err != nil {
		return now, uuid.Nil, fmt.Errorf(format, ErrUUID, err)
	}

	return now, uid, nil
}
