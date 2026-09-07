package cache_test

import (
	"testing"
	"time"

	"github.com/Defacto2/server/handler/cache"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage was great at 80%+

func TestCache(t *testing.T) {
	t.Parallel()

	ct := cache.Test
	be.Equal(t, ct.String(), "test")

	got, err := ct.Path()
	be.Err(t, err, nil)
	be.True(t, len(got) > 0)

	const key = "foo"
	const immediate = 0
	err = ct.Write(key, "bar", immediate)
	be.Err(t, err, nil)

	const one = 1 * time.Millisecond
	err = ct.Write(key, "bar", one)
	be.Err(t, err, nil)

	err = ct.WriteNoExpire(key, "bar")
	be.Err(t, err, nil)

	s, err := ct.Read(key)
	be.Err(t, err, nil)
	be.Equal(t, s, "bar")

	err = ct.WriteNoExpire(key, "xxx")
	be.Err(t, err, nil)

	s, err = ct.Read(key)
	be.Err(t, err, nil)
	be.Equal(t, s, "xxx")

	err = ct.Delete(key)
	be.Err(t, err, nil)
}
