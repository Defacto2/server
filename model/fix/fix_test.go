package fix_test

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model/fix"
	"github.com/nalgeon/be"
)

// checked in Aug 26, test coverage was fine at around 60%+

func TestRepairRunInvalid(t *testing.T) {
	t.Parallel()

	const invalid = -2
	r := fix.Repair(invalid)
	got := r.Run(t.Context(), nil, nil, nil)
	be.Err(t, got)

	sl := slog.Default()
	r = fix.Releaser
	got = r.Run(t.Context(), sl, nil, nil)
	be.Err(t, got)
}

func TestRepairArtifactsRun(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	tx := testutil.Tx(t)
	sl := slog.Default()

	got := fix.Artifacts.Run(t.Context(), sl, nil, tx)
	be.Err(t, got, nil)

	got = fix.Releaser.Run(t.Context(), sl, db, tx)
	be.Err(t, got, nil)
}

func TestGetNumericSuffix(t *testing.T) {
	t.Parallel()

	got, err := fix.NumSuffix(context.TODO(), nil)
	be.Err(t, err)
	be.True(t, got != nil)
	be.True(t, got.Count == 0)
	be.True(t, len(got.Files) == 0)

	_, err = fix.NumSuffix(t.Context(), nil)
	be.Err(t, err)

	db := testutil.DB(t)
	got, err = fix.NumSuffix(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, got != nil)
	be.True(t, got.Count >= 0)
}

func TestSyncFilesIDSeqNoDB(t *testing.T) {
	t.Parallel()

	got := fix.SyncFilesIDSeq(nil)
	be.Err(t, got)

	db := testutil.DB(t)

	got = fix.SyncFilesIDSeq(db)
	be.Err(t, got, nil)
}

func TestRepairString(t *testing.T) {
	t.Parallel()

	const invalid = fix.Repair(-10)
	tests := []struct {
		r    fix.Repair
		want string
	}{
		{fix.None, "skip"},
		{fix.Artifacts, "on all artifacts"},
		{fix.Releaser, "on the releasers"},
		{invalid, "error, unknown"},
	}
	for _, tt := range tests {
		got := tt.r.String()
		be.Equal(t, got, tt.want)
	}
}

func TestFileModelCreation(t *testing.T) {
	t.Parallel()

	const id = 123
	f := &models.File{ID: id}
	be.Equal(t, f.ID, int64(id))
}

func TestReplacement(t *testing.T) {
	t.Parallel()

	replacements := fix.Replacements()

	for old, val := range replacements {
		be.True(t, old != string(val))
	}
}

func TestReplacementsCase(t *testing.T) {
	t.Parallel()

	replacements := fix.Replacements()
	tests := []struct {
		input    string
		expected string
	}{
		{"lowercase", "LOWERCASE"},
		{"UPPERCASE", "UPPERCASE"},
		{"MixedCase", "MIXEDCASE"},
	}

	for _, tc := range tests {
		lookup := strings.ToUpper(tc.input)
		if val, exists := replacements[lookup]; exists {
			be.True(t, string(val) == tc.expected)
		}
	}
}
