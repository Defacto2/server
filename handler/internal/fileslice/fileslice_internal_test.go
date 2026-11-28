//nolint:paralleltest
package fileslice

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/tags"
	"github.com/nalgeon/be"
)

var errStubQuery = errors.New("stub query error")

type emptyRows struct {
	cols []string
	data [][]driver.Value
	idx  int
}

func (r *emptyRows) Columns() []string { return r.cols }
func (r *emptyRows) Close() error      { return nil }
func (r *emptyRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.idx])
	r.idx++
	return nil
}

type fnConn struct {
	queryFn func(query string) (driver.Rows, error)
}

func (c *fnConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *fnConn) Close() error                        { return nil }
func (c *fnConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }
func (c *fnConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	return c.queryFn(query)
}

type fnConnector struct {
	queryFn func(query string) (driver.Rows, error)
}

func (c fnConnector) Connect(context.Context) (driver.Conn, error) {
	return &fnConn{queryFn: c.queryFn}, nil
}
func (c fnConnector) Driver() driver.Driver { return nil }

func TestCounterOffline(t *testing.T) {
	// Not parallel: resets package-level statsCache before and after the test
	// so parallel live-DB tests see a clean cache.
	resetCache := func() {
		statsMu.Lock()
		statsCache = Stats{}
		statsCacheTime = time.Time{}
		statsMu.Unlock()
	}
	resetCache()
	t.Cleanup(resetCache)

	sl := logs.Discard()
	_, err := Counter(t.Context(), nil, nil)
	be.Err(t, err)

	errDB := sql.OpenDB(fnConnector{
		queryFn: func(string) (driver.Rows, error) { return nil, errStubQuery },
	})
	defer func() { _ = errDB.Close() }()

	_, err = Counter(t.Context(), sl, errDB)
	be.Err(t, err, errStubQuery)

	uris := tags.URIs()
	okDB := sql.OpenDB(fnConnector{
		queryFn: func(query string) (driver.Rows, error) {
			if strings.Contains(query, "BETWEEN 1980 AND 2010") {
				return &emptyRows{
					cols: []string{"date_issued_year", "platform", "count", "sum"},
					data: [][]driver.Value{{int64(1995), uris[tags.DOS], int64(3), int64(300)}},
				}, nil
			}
			return &emptyRows{
				cols: []string{"section", "platform", "count", "sum", "min", "max"},
				data: [][]driver.Value{{uris[tags.Nfo], uris[tags.DOS], int64(3), int64(300), int64(1995), int64(1995)}},
			}, nil
		},
	})
	defer func() { _ = okDB.Close() }()

	got, err := Counter(t.Context(), sl, okDB)
	be.Err(t, err, nil)
	be.Equal(t, got.Record.Count, 3)

	// Second call must hit the cache even when passed a failing DB.
	cached, err := Counter(t.Context(), sl, errDB)
	be.Err(t, err, nil)
	be.Equal(t, cached.Record.Count, 3)
}

func TestDispatchInvariants(t *testing.T) {
	t.Parallel()

	be.Equal(t, len(uriStrings), int(_maxURI))
	for i := 1; i <= int(LastURI); i++ {
		uri := URI(i)
		_, inDispatch := recordDispatch[uri]
		be.True(t, inDispatch || uri == Sensenstahl)
	}
}
