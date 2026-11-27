package cache_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/cache"
	"github.com/nalgeon/be"
	"github.com/rosedblabs/rosedb/v2"
)

func TestString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c    cache.Cache
		want string
	}{
		{"pouet vote", cache.PouetVote, "pouet"},
		{"pouet production", cache.PouetProduction, "pouetproduction"},
		{"demozoo production", cache.DemozooProduction, "demozooproduction"},
		{"test", cache.Test, "test"},
		{"negative", cache.Cache(-1), ""},
		{"out of range", cache.Cache(99), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			be.Equal(t, tt.c.String(), tt.want)
		})
	}
}

func TestPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)

	// First call creates the directory.
	got, err := cache.Test.Path()
	be.Err(t, err, nil)
	be.True(t, strings.HasPrefix(got, tmp))
	info, err := os.Stat(got)
	be.Err(t, err, nil)
	be.True(t, info.IsDir())

	// Second call hits the existing-directory branch.
	got2, err := cache.Test.Path()
	be.Err(t, err, nil)
	be.Equal(t, got2, got)

	// Invalid cache values fail without panicking.
	_, err = cache.Cache(-1).Path()
	be.Err(t, err, cache.ErrInvalid)
	_, err = cache.Cache(99).Path()
	be.Err(t, err, cache.ErrInvalid)

	// Non-directory parent triggers an os.Stat error.
	blocker := filepath.Join(tmp, "not-a-dir")
	err = os.WriteFile(blocker, []byte("x"), 0o600)
	be.Err(t, err, nil)
	t.Setenv("TMPDIR", blocker)
	_, err = cache.Test.Path()
	be.Err(t, err)
}

func TestCache(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())

	ct := cache.Test
	const key = "foo"

	// Immediate expiry (ttl == 0) should not be readable afterward.
	err := ct.Write(key, "expired", 0)
	be.Err(t, err, nil)
	_, err = ct.Read(key)
	be.Err(t, err, rosedb.ErrKeyNotFound)

	// Positive TTL is readable before expiry.
	err = ct.Write(key, "alive", time.Minute)
	be.Err(t, err, nil)
	s, err := ct.Read(key)
	be.Err(t, err, nil)
	be.Equal(t, s, "alive")

	// Short TTL expires after the duration elapses.
	err = ct.Write("short", "temp", 15*time.Millisecond)
	be.Err(t, err, nil)
	time.Sleep(25 * time.Millisecond)
	_, err = ct.Read("short")
	be.Err(t, err, rosedb.ErrKeyNotFound)

	// WriteNoExpire writes and overwrites values.
	err = ct.WriteNoExpire(key, "bar")
	be.Err(t, err, nil)
	s, err = ct.Read(key)
	be.Err(t, err, nil)
	be.Equal(t, s, "bar")

	err = ct.WriteNoExpire(key, "xxx")
	be.Err(t, err, nil)
	s, err = ct.Read(key)
	be.Err(t, err, nil)
	be.Equal(t, s, "xxx")

	// Delete removes the key so subsequent Read fails with key in error message.
	err = ct.Delete(key)
	be.Err(t, err, nil)
	_, err = ct.Read(key)
	be.Err(t, err, rosedb.ErrKeyNotFound)
	be.True(t, strings.Contains(err.Error(), `key "foo" cache read from rosedb`))

	// Empty keys return ErrKeyIsEmpty across all operations.
	be.Err(t, ct.Write("", "v", time.Second), rosedb.ErrKeyIsEmpty)
	be.Err(t, ct.WriteNoExpire("", "v"), rosedb.ErrKeyIsEmpty)
	_, err = ct.Read("")
	be.Err(t, err, rosedb.ErrKeyIsEmpty)
	err = ct.Delete("")
	be.Err(t, err, rosedb.ErrKeyIsEmpty)
	be.True(t, strings.Contains(err.Error(), `key "" cache delete from rosedb`))
}

func TestCachePathErrors(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "not-a-dir")
	err := os.WriteFile(blocker, []byte("x"), 0o600)
	be.Err(t, err, nil)
	t.Setenv("TMPDIR", blocker)

	ct := cache.Test
	be.Err(t, ct.Write("k", "v", time.Second))
	be.Err(t, ct.WriteNoExpire("k", "v"))
	_, err = ct.Read("k")
	be.Err(t, err)
	be.Err(t, ct.Delete("k"))
}
