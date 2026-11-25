package fix_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model/fix"
	"github.com/nalgeon/be"
)

// checked in Aug 26, test coverage was fine at around 60%+

func TestRepairInvalid(t *testing.T) {
	t.Parallel()

	const invalid = -2
	r := fix.Repair(invalid)
	got := r.Run(t.Context(), nil, nil)
	be.Err(t, got)

	sl := logs.Discard()
	r = fix.Releaser
	got = r.Run(t.Context(), sl, nil)
	be.Err(t, got)
}

func TestRepairRun(t *testing.T) {
	t.Parallel()

	tx := testutil.Tx(t)
	sl := logs.Discard()

	got := fix.Artifacts.Run(t.Context(), sl, tx)
	be.Err(t, got, nil)

	got = fix.Releaser.Run(t.Context(), sl, tx)
	be.Err(t, got, nil)
}

func TestNumSuffix(t *testing.T) {
	t.Parallel()

	got, err := fix.NumSuffix(t.Context(), nil)
	be.Err(t, err)
	be.True(t, got != nil)
	be.Equal(t, got.Count, 0)
	be.Equal(t, len(got.Files), 0)

	db := testutil.DB(t)
	got, err = fix.NumSuffix(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, got != nil)
	be.Equal(t, got.Count, int64(len(got.Files)))
}

func TestSyncIDSeq(t *testing.T) {
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

func TestReplacements(t *testing.T) {
	t.Parallel()

	replacements := fix.Replacements()
	be.True(t, len(replacements) > 0)

	for old, val := range replacements {
		be.True(t, old != "")
		be.Equal(t, old, strings.ToUpper(old))
		be.True(t, strings.TrimSpace(string(val)) != "")
		be.True(t, old != string(val))
	}
}

func TestIndexes(t *testing.T) {
	t.Parallel()

	_, err := fix.PlanIndexes(t.Context(), nil)
	be.Err(t, err)
	be.Err(t, fix.RunIndexes(t.Context(), nil, nil, true))
	be.Err(t, fix.Indexes(t.Context(), nil, nil))

	db := testutil.DB(t)
	plan, err := fix.PlanIndexes(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, len(plan.Existing)+len(plan.Create) >= 9)

	err = fix.RunIndexes(t.Context(), logs.Discard(), db, true)
	be.Err(t, err, nil)
}
