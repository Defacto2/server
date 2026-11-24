# SQL & Database Architecture Review

*Reviewed: November 2026*

This document reviews all SQL usage across the Defacto2 server codebase and evaluates three potential changes:

1. Upgrading from PostgreSQL 16 to PostgreSQL 19
2. Migrating from PostgreSQL to SQLite
3. Replacing SQLBoiler with [Bob](https://bob.stephenafamo.com) or another Go SQL tool

---

## Current SQL Profile

All SQL usage in the application lives in:

- `internal/postgres/models/` (SQLBoiler-generated models for the `files` table)
- `internal/postgres/sql.go` and `internal/postgres/postgres.go` (raw PostgreSQL queries and connection helpers)
- `model/`, `model/querymod/`, `model/fix/`, and `model/html3/` (ORM queries, filters, aggregations, and maintenance routines)
- `internal/tags/tags.go` (section and platform group-count queries)
- `internal/config/fixer.go`, `repair.go`, and `fixarc`/`fixarj`/`fixlha`/`fixzip` (startup checks and archive repair queries)

### Key Characteristics

- **Single-table schema:** The entire database consists of one table (`files`, ~45 columns, ~55k+ rows) with no foreign keys or multi-table joins.
- **Read-heavy workload:** Catalog browsing, pagination, tag/releaser/scener aggregations, and search, with occasional admin edits/uploads, startup repair jobs, and an in-process RoseDB key/value cache.
- **Split query layer:** Simple CRUD and ~55 category/platform filter models in `model/filter.go` (2,335 lines) and `model/summary.go` (573 lines) use SQLBoiler v4 (`qm.QueryMod`), while all complex queries bypass SQLBoiler via raw SQL strings (`queries.Raw`, `qm.SQL`, `exec.QueryContext`).
- **Heavy reliance on PostgreSQL-specific SQL:**
  - `CROSS JOIN LATERAL (VALUES (group_brand_for), (group_brand_by)) AS T(releaser)` in `internal/postgres/sql.go` and `model/querymod/querymod.go` (`OneOffsExpr`).
  - `DISTINCT ON (upper(scener))` in `internal/postgres/sql.go`.
  - Full-text search (`to_tsvector('english', COALESCE(record_title, '') || ' ' || COALESCE(comment, '')) @@ plainto_tsquery('english', ?)`) in `model/artifacts.go` and `model/summary.go`.
  - POSIX regex operators `~` / `!~` with `\M` word-end anchors (`BBS\M`, `FTP\M`, numeric suffix detection, UUID validation) and `regexp_replace` in `internal/postgres/sql.go` and `model/fix/fix.go`.
  - `SIMILAR TO`, `ILIKE`, `ILIKE ANY(ARRAY[...])`, `BTRIM`, `strpos`, `right`, and `::text` casts.
  - PostgreSQL internals: `pg_stat_activity`, `pg_settings`, `nextval`/`setval` on `files_id_seq`, and `VACUUM ANALYZE files`.

---

## 1. Upgrading PostgreSQL 16 → PostgreSQL 19

**Verdict: Low-effort, worthwhile upgrade — though PostgreSQL 18 is the safer sweet spot right now.**

### Compatibility

The driver (`github.com/jackc/pgx/v5` v5.13.0) and all SQL queries in the codebase are forward-compatible with PostgreSQL 17, 18, and 19 with no breaking changes.

### Benefits Relevant to This App

- **PostgreSQL 17:** Overhauled `VACUUM` memory management (faster `VACUUM ANALYZE files` in `model/fix/fix.go`), faster CTE execution, and native incremental backups.
- **PostgreSQL 18:**
  - **Asynchronous I/O (`io_uring` / worker pool):** The biggest real-world win for this workload. Queries using `ILIKE '%...%'`, `to_tsvector`, `CROSS JOIN LATERAL` releaser rollups, and `LOWER(BTRIM(...))` stats perform sequential and bitmap heap scans, which AIO accelerates significantly.
  - **B-Tree skip scan:** Helps multi-column B-tree indexes accelerate queries even when leading columns are omitted from `WHERE`.
  - **Virtual generated columns:** Useful for cleanly expressing computed values like normalized `LOWER(BTRIM(...))` or `tsvector` columns without extra storage.
  - **Native `uuidv7()`:** Available in SQL (though Go already generates UUIDv7 in `model/insert.go`).
- **PostgreSQL 19 (released November 2026):**
  - **`REPACK (CONCURRENTLY)`:** Online table bloat cleanup without exclusive locks.
  - **Parallel autovacuum** and improved observability views.
  - **`pg_plan_advice`** and `INSERT ... ON CONFLICT DO SELECT`.

### Caveats

- Because PostgreSQL 19 is a brand-new `.0` release, **PostgreSQL 18** provides the main performance gains (Async I/O, skip scan, improved VACUUM) with a full year of patch releases behind it.
- For a single ~66k-row table, **targeted indexing** delivers a far larger latency improvement than a major version bump alone — now implemented via the `index` (`i`) CLI subcommand (see [PostgreSQL Indexing & Benchmark Results](#postgresql-indexing--benchmark-results) below).

### Upgrading PG 16 → PG 18 on Ubuntu 24.04 (Noble)

Ubuntu 24.04 only ships PostgreSQL 16 in its default apt repositories, so upgrading to PostgreSQL 18 requires adding the official PGDG apt repo (built-in helper script in `postgresql-common`) and migrating via `pg_upgradecluster`. PG 16 remains supported with security patches until November 2028, so there is no security urgency.

1. **Take a safety backup:**
   ```bash
   sudo -u postgres pg_dump -Fc defacto2_ps > ~/defacto2_ps_pg16_$(date +%F).dump
   ```

2. **Add the official PGDG apt repo and install PostgreSQL 18:**
   ```bash
   sudo apt update
   sudo apt install -y postgresql-common
   sudo /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh
   sudo apt install -y postgresql-18
   ```
   *(Installs PG 18 alongside PG 16; the new empty PG 18 cluster starts on port `5433` while PG 16 stays untouched on `5432`.)*

3. **Stop Defacto2 and migrate the cluster:**
   ```bash
   sudo systemctl stop defacto2

   # Remove the empty default PG 18 cluster so pg_upgradecluster can migrate 16 -> 18 onto port 5432
   sudo pg_dropcluster --stop 18 main

   # Migrate PG 16 -> PG 18 (default dump/restore takes seconds at ~55k rows and enables PG 18's default data checksums)
   sudo pg_upgradecluster 16 main
   ```

4. **Verify and restart Defacto2:**
   ```bash
   pg_lsclusters
   sudo -u postgres psql -d defacto2_ps -c "SELECT version(), COUNT(*) FROM files;"
   sudo systemctl start defacto2
   ```

5. **Clean up PG 16 once verified:**
   ```bash
   sudo pg_dropcluster 16 main
   sudo apt purge -y postgresql-16 postgresql-client-16
   sudo apt autoremove -y
   ```

### Upgrading Local Docker (`postgres16` → `postgres18`)

Major PostgreSQL versions have incompatible on-disk data directories, so Docker upgrades require a **new `postgres:18` container** with a fresh volume and a `pg_dump` / `pg_restore` (takes a few seconds at ~55k rows and preserves `files_id_seq` automatically).

1. **Dump from `postgres16` and stop it:**
   ```bash
   docker exec postgres16 pg_dump -U root -Fc defacto2_ps > defacto2_ps_pg16.dump
   docker stop postgres16
   ```

2. **Start a new `postgres:18` container on port `5432`:**
   ```bash
   docker run -d \
     --name postgres18 \
     -e POSTGRES_USER=root \
     -e POSTGRES_PASSWORD=example \
     -e POSTGRES_DB=defacto2_ps \
     -p 5432:5432 \
     -v postgres18_data:/var/lib/postgresql \
     postgres:18
   ```
   > **Note:** Starting with `postgres:18`, the official Docker image mounts persistent volumes at `/var/lib/postgresql` (instead of `/var/lib/postgresql/data`), storing versioned data under `/var/lib/postgresql/18/docker`.

3. **Restore the dump and verify:**
   ```bash
   docker exec -i postgres18 pg_restore -U root -d defacto2_ps --no-owner < defacto2_ps_pg16.dump
   docker exec -it postgres18 psql -U root -d defacto2_ps -c "SELECT version(), COUNT(*) FROM files;"
   ```
   *(Once verified, remove the old container with `docker rm postgres16`.)*

---

## 2. Changing to SQLite

**Verdict: Architecturally a strong fit for Defacto2's deployment, with a moderate SQL rewrite cost.**

### Pros

- **Workload fit:** Single table (~55k rows), read-heavy, single-node systemd deployment (`init/defacto2.service`) is SQLite's sweet spot.
- **Zero network overhead:** In-process reads with sub-millisecond latency.
- **Simpler operations:** No PostgreSQL daemon or Docker container to manage; trivial single-file backups.
- **Self-contained tests:** Currently, database tests silently skip when Postgres is unavailable (`internal/testutil/database.go`). With SQLite, tests can run everywhere against an in-memory or temp-file DB.
- **Concurrency:** WAL mode + `busy_timeout` easily handles concurrent readers with occasional admin writes on a single server.

### Required SQL & Driver Changes

Migrating off PostgreSQL requires rewriting the Postgres-specific SQL in `internal/postgres/sql.go`, `model/fix/fix.go`, `model/artifacts.go`, and `model/summary.go`:

| PostgreSQL Feature | SQLite Equivalent |
| :--- | :--- |
| `CROSS JOIN LATERAL (VALUES (col1), (col2))` | `UNION ALL` CTE or subquery |
| `DISTINCT ON (upper(scener))` | `GROUP BY` or `ROW_NUMBER() OVER (PARTITION BY ...)` |
| `to_tsvector` / `plainto_tsquery` | **FTS5** virtual table + sync triggers (often faster than unindexed `tsvector`, but adds trigger maintenance) |
| POSIX regex `~` / `!~`, `regexp_replace` | No built-in regex engine; register a Go `REGEXP` UDF or rewrite simple patterns (`BBS\M`, `FTP\M`) with `LIKE`/Go filtering |
| `SIMILAR TO` | Rewrite with `LIKE` / `OR` or Go-side matching |
| `ILIKE` / `ILIKE ANY(ARRAY[...])` | `LIKE` (case-insensitive for ASCII by default) or `COLLATE NOCASE` + `OR` |
| `BTRIM`, `strpos`, `right(s, 1)` | `TRIM`, `INSTR`, `SUBSTR(s, -1)` |
| `pg_stat_activity`, `files_id_seq`, `VACUUM ANALYZE files` | Drop or replace with `PRAGMA`s and `VACUUM; ANALYZE;` |

You would also need to choose a Go driver: pure-Go `modernc.org/sqlite` (no CGO, seamless GoReleaser cross-compilation, slightly slower) vs. CGO-based `mattn/go-sqlite3` (or WASM `ncruces/go-sqlite3`).

### Concurrency & Single-Writer Gotchas in This Implementation

SQLite allows only **one writer at a time** (database-level write lock, rather than row-level locks). In WAL mode (`PRAGMA journal_mode=WAL`), **unlimited readers run concurrently** without blocking the writer, so Defacto2's read-heavy traffic won't hit write throughput limits. However, four specific patterns in the current server code must be adjusted for SQLite:

1. **[Fixed] Long-held background transaction during HTTP downloads (`handler/htmx/submitprod.go`, `handler/routerlock.go`, `handler/app/remote/`):**
   Previously `fetchRemote` and `/get/demozoo/download/:unid/:id` opened a transaction *before* downloading remote files from Demozoo/Pouet. Now both pass `db` directly (`boil.ContextExecutor`) so no transaction is held open across network I/O.
2. **[Fixed] `VACUUM` executed while a write transaction is open (`internal/config/fixer.go`, `model/fix/fix.go`):**
   `databaseFix` now commits `tx` first and then calls `fix.Optimize(ctx, db)` after commit, and the unused `db` parameter was removed from `fix.Repair.Run`.
3. **HTMX concurrent edits & `DEFERRED` transaction upgrade deadlocks (`handler/router.go:29-51`, `model/update.go`):**
   Each HTMX editor field triggers its own `BeginTx` request, and `model/update.go` runs `SELECT` then `UPDATE` inside that transaction. SQLite's default `BEGIN DEFERRED` starts as a read transaction and upgrades on `UPDATE`; if two field saves arrive simultaneously, both try to upgrade and SQLite immediately returns `SQLITE_BUSY` without honoring `busy_timeout`.
   - *Fix:* Use `BEGIN IMMEDIATE` (e.g., `_txlock=immediate` in the DSN) for write transactions so writers queue cleanly.
4. **Connection pooling with parallel `errgroup` reads (`model/count.go:50-74`, `handler/handler.go:545`):**
   Setting `SetMaxOpenConns(1)` on a single shared `*sql.DB` would serialize parallel `errgroup` reads and risk deadlocks if `db` is used while a `tx` is open.
   - *Fix:* Use the **split-pool pattern** pointing at the same WAL-mode `.db` file:
     - `readDB`: `SetMaxOpenConns(runtime.NumCPU())` for all `SELECT` / `errgroup` queries
     - `writeDB`: `SetMaxOpenConns(1)` + `busy_timeout=5000` + `BEGIN IMMEDIATE` for all writes

### Portability & Preservation Angle

For a digital preservation project like Defacto2, SQLite unlocks a self-contained, zero-setup distribution model:

- **Zero DB setup for contributors and mirrors:** Cloning the repo (or downloading a release) and running `./defacto2-server` works immediately — no PostgreSQL install, Docker container, credentials, or sequence troubleshooting (`files_id_seq`).
- **Archival longevity:** SQLite is a Library of Congress recommended archival format; a `.db` file remains readable by standard tools decades from now without requiring a running database server.
- **Single-binary + single-file distribution:** `server.go` already embeds `public/` and `view/` via `embed.FS`. Paired with a pure-Go SQLite driver (`modernc.org/sqlite`), the app runs on Linux, macOS, or Windows with zero external dependencies.
- **Read-only mirrors:** Combines naturally with `D2_READ_ONLY` using SQLite's immutable read-only mode (`?mode=ro&immutable=1`) for offline archives or read-only mirrors.
- **GitHub distribution notes:** A 55k-row table (with `file_zip_content` + FTS5 index) will be roughly 30–80 MB uncompressed. Rather than committing a large binary `.db` directly into Git history (near GitHub's 50 MB soft / 100 MB hard limits), publish snapshots as **GitHub Release assets** (or keep a small seed `.db` in Git + full `.db` in Releases). Run `PRAGMA wal_checkpoint(TRUNCATE); VACUUM;` before publishing snapshots so they are clean single files without `-wal`/`-shm` sidecars.

### Bottom Line

Worth doing **if** the goal is operational simplicity, portability, and long-term preservation (dropping the Postgres service, single-file backups, zero-setup local runs and tests). Going **SQLite-only** is cleaner than dual-supporting Postgres + SQLite due to the search (`tsvector` vs FTS5) and regex differences. Not necessary purely for performance if PostgreSQL is already running smoothly.

---

## 3. Swapping Out SQLBoiler (Bob vs. Alternatives)

**Verdict: Clear value in moving off SQLBoiler.**

### Why Move Off SQLBoiler

- **Maintenance-only status:** SQLBoiler is in strict maintenance mode with no new feature development.
- **Extra friction & dependencies:** Requires a live PostgreSQL server to run `go generate` (`init/.sqlboiler.toml`) and pulls in legacy packages (`aarondl/null/v8`, `friendsofgo/errors`, `aarondl/strmangle`).
- **Underutilized:** With only one table (`files`), the codebase does not use SQLBoiler's primary strength (multi-table relationship eager-loading).
- **Already bypassed:** Every non-trivial query already drops down to raw SQL string concatenation (see the trailing-space conventions in `internal/postgres/sql.go`).

### Comparison of Alternatives

| Tool | Best For | Pros | Cons |
| :--- | :--- | :--- | :--- |
| **[Bob](https://bob.stephenafamo.com)** | Closest drop-in successor to SQLBoiler | Built as SQLBoiler's spiritual successor (`QueryMod`s, soft deletes, PG & SQLite support). Its typed query builder natively handles `LATERAL`, CTEs, and `DISTINCT ON`, *and* its `.sql` query codegen composes directly with dynamic `QueryMod`s. | Fairly large API and codegen footprint for a 1-table schema. |
| **`sqlc`** | SQL-first simplicity & native `pgx/v5` | Write plain `.sql` files and get minimal, compile-time-checked Go code with native `pgx/v5` support (no `database/sql` wrapper). Eliminates `models/files.go` and `aarondl/null`. Supports Postgres and SQLite. | Dynamic queries (`OnlyDescriptions`, `OnlyFilenames`, `OnlyUniqueIDs`, `html3` dynamic `ORDER BY`) need `ANY($1::text[])` / `CASE` patterns or small helpers. |
| **Plain `pgx/v5` (`pgx.CollectRows`)** | Zero codegen for a 1-table DB | With only one table, a single hand-written `File` struct + `pgx.CollectRows` removes `go generate`, `.sqlboiler.toml`, and 5 dependencies. | No compile-time SQL validation (relies on the test suite). |
| **`go-jet`** | Heavy dynamic query building | Compile-time type-safe Go SQL DSL supporting both Postgres and SQLite; fits `model/filter.go` and `model/html3/html3order.go` well. | Custom DSL to learn; more verbose than raw SQL for Postgres-specific expressions. |

### Recommendation

- Choose **Bob** if you like SQLBoiler's model + composable `QueryMod` workflow and want type-safe coverage of `internal/postgres/sql.go` (plus an easier future path to SQLite).
- Choose **`sqlc`** if you want to simplify the data layer, use `pgx/v5` natively, and reduce generated code — especially if paired with collapsing the repetitive filter structs in `model/filter.go` into a data-driven map/helper.

### Other Go SQL Utilities to Consider

Beyond Bob, `sqlc`, `go-jet`, and plain `pgx/v5`, several other Go utilities fit a single-table, read-heavy codebase:

#### Lightweight Query Builders (No Codegen)

Good if you want to drop `go generate` completely while still building dynamic `WHERE` / `ORDER BY` clauses cleanly (`model/artifacts.go`, `model/filter.go`, `model/html3/html3order.go`).

- **[`uptrace/bun`](https://bun.uptrace.dev):** SQL-first ORM/builder with first-class PostgreSQL and SQLite dialects, built-in soft deletes (`deletedat`), struct scanning, and expressive SQL-like chaining. Closest zero-codegen equivalent to how SQLBoiler is used today.
- **[`huandu/go-sqlbuilder`](https://github.com/huandu/go-sqlbuilder):** Actively maintained pure SQL query builder (modern alternative to the mostly-dormant `Masterminds/squirrel`). Handles Postgres `$1` and SQLite `?` placeholders automatically, struct-to-column mapping, and dynamic `OR`/`AND` chains without taking over the DB driver.
- **[`doug-martin/goqu`](https://github.com/doug-martin/goqu):** Expressive SQL builder with Postgres/SQLite dialects, CTEs, and struct scanning, though less actively developed than `bun` or `go-sqlbuilder`.

#### Struct Scanners & `database/sql` Helpers

Good if you write your own SQL or pair with a pure query builder like `go-sqlbuilder`.

- **[`jmoiron/sqlx`](https://github.com/jmoiron/sqlx):** Battle-tested `database/sql` wrapper (`SelectContext`, `GetContext`, `NamedExec`, `sqlx.In` — replaces manual placeholder loops like in `model/fix/duplicates.go`). Works identically with `pgx/v5/stdlib` and SQLite drivers.
- **[`georgysavva/scany`](https://github.com/georgysavva/scany):** Minimalist scanner supporting both `database/sql` and native `pgx/v5` (`pgxscan.Select`).
- **[`stephenafamo/scan`](https://github.com/stephenafamo/scan):** Standalone scanning package extracted from Bob; works with any `database/sql` driver if you want Bob's type-safe scanning without adopting the full ORM.
- *Note:* If staying on Postgres and switching from `database/sql` to native `pgx/v5` (`pgxpool`), `pgx.CollectRows(..., pgx.RowToStructByName[File])` is already built in.

#### Type-Safe Builder + Optional Codegen

- **[`bokwoon95/sq`](https://github.com/bokwoon95/sq):** Type-safe Go SQL query builder (Postgres and SQLite) with optional table-struct codegen, CTEs, and `LATERAL` support. Smaller community than Bob or `go-jet`, but clean SQL-like ergonomics.

#### Complementary Schema & Migration Tools

Currently the schema lives only in the live Postgres database (no tracked SQL migrations). Pairing `sqlc`, Bob, or SQLite with a migration tool makes index changes and test DB setup reproducible:

- **[`pressly/goose`](https://github.com/pressly/goose):** Embeddable Go SQL migrations (`embed.FS`), supports both Postgres and SQLite.
- **[`golang-migrate/migrate`](https://github.com/golang-migrate/migrate)** or **[`atlasgo.io`](https://atlasgo.io):** Versioned or declarative diff-based schema migrations.

#### Tools to Skip for This App

- **GORM:** Heavy reflection/magic that fights raw Postgres SQL (`LATERAL`, `tsvector`, `DISTINCT ON`) and adds unnecessary complexity for a single table.
- **Ent (`entgo.io`):** Designed for complex multi-table entity graphs; overkill for one `files` table.

#### Top Picks Beyond Bob & `sqlc`

- **Zero codegen + soft deletes + Postgres/SQLite support:** `uptrace/bun`
- **Maximum simplicity & full SQL control:** `huandu/go-sqlbuilder` + `sqlx` (or native `pgx/v5` `CollectRows` if staying on Postgres)

#### Dependency Impact (`go.mod`) of Moving to `uptrace/bun`

Removing SQLBoiler immediately drops **4 direct** (`aarondl/sqlboiler/v4`, `aarondl/null/v8`, `aarondl/strmangle`, `friendsofgo/errors`) and **4+ indirect** (`aarondl/inflect`, `aarondl/randomize`, `gofrs/uuid`, `spf13/cast`) dependencies — plus `jackc/pgx/v5` and its 3 indirects (`pgpassfile`, `pgservicefile`, `puddle/v2`) if also leaving PostgreSQL.

`uptrace/bun` adds **2 direct** modules (`uptrace/bun` + either `pgdialect` or `sqlitedialect`) and ~4 small indirects (`jinzhu/inflection`, `puzpuzpuz/xsync`, `tmthrgd/go-hex`, `vmihailenco/msgpack/v5`):

| Setup | Direct DB Deps | Indirect DB Deps | Notes |
| :--- | :--- | :--- | :--- |
| **Current (SQLBoiler + pgx)** | 5 | 7 | Also requires `sqlboiler` CLI + live DB for `go generate` |
| **Bun + pgx (Postgres)** | 3 | ~7 | Drops 2 direct deps, no codegen |
| **Bun + `mattn/go-sqlite3`** | 3 | ~4 | Smallest dep tree, requires CGO |
| **Bun + `modernc.org/sqlite`** | 3 | ~12 | Pure Go (easy cross-compile), pulls in `modernc.org/*` helpers |
| **`sqlc` + driver only** | 1 | 0–8 | `sqlc` adds **0 runtime dependencies** to `go.mod` |

> **Tip:** If using Bun with SQLite, import `modernc.org/sqlite` or `mattn/go-sqlite3` directly rather than `uptrace/bun/driver/sqliteshim`, which pulls *both* SQLite drivers into `go.mod`.

---

## Expected File Size & Performance Impact

### Compiled Binary Size

The compiled `server` binary (~56 MB) is primarily driven by `embed.FS` assets (`public/` and `view/`), `google.golang.org/api`, `golang.org/x/*`, and Go runtime/symbols.

- **SQLBoiler → `uptrace/bun` (PostgreSQL):** Negligible change (~±0.5–1 MB). Removing generated `models/files.go` + SQLBoiler packages is roughly offset by `bun`'s reflection/builder code.
- **SQLBoiler → `sqlc` (PostgreSQL):** Small reduction (~1–2 MB), since `sqlc` has zero runtime library overhead.
- **PostgreSQL (`pgx/v5`) → SQLite:**
  - Pure-Go `modernc.org/sqlite`: **+5–8 MB** (embeds the transpiled SQLite C engine inside the Go binary).
  - CGO `mattn/go-sqlite3`: **+1–2 MB** (or ~0 MB if dynamically linked to system `libsqlite3`).

### Database Size on Disk

- **SQLite is typically 30–50% smaller** than PostgreSQL's `PGDATA` footprint for the same ~55k-row `files` table, since SQLite uses compact variable-length encoding without Postgres's per-row MVCC headers (23+ bytes/row), TOAST tables, or WAL segment overhead. Adding an FTS5 index adds back some size, though still compact after `VACUUM`.

### Runtime Performance & Memory

- **Server RAM:** Dropping the PostgreSQL daemon saves **50–250+ MB** of resident RAM; in-process SQLite uses a configurable page cache (typically 16–32 MB).
- **Simple queries (`One`, `ByKey`, `Count`, `Exists`, tag filters):** SQLite is generally **faster** on a single host because there is zero TCP/Unix-socket round-trip or wire serialization (microseconds vs. ~0.5–2 ms).
- **Builder overhead:** `uptrace/bun` uses reflection for scanning (similar to SQLBoiler's `Bind`, slightly faster than `boil.Infer()` on writes); `sqlc` + native `pgx/v5` avoids reflection entirely (~10–30% less Go CPU/allocations per query, though usually overshadowed by DB I/O).
- **Complex queries & search:**
  - **Full-text search:** SQLite **FTS5** and PostgreSQL's `GIN` expression index (`idx_files_desc_fts`, ~3 ms on 66k rows) achieve comparable sub-5 ms search latency.
  - **Heavy scans & aggregations:** PostgreSQL (especially PG 18+ with Async I/O and parallel workers) handles large full-table regex/aggregation scans faster under high concurrency than single-threaded SQLite. At ~55k rows with RoseDB caching, both feel instant (<5–20 ms) once properly indexed.

---

## SQL Issues & Quick Wins Spotted During Review

Regardless of database or query-builder changes, these existing issues were identified and resolved:

1. **[Fixed] Parameter index bug in `Summary.ByDescription` (`model/summary.go:44-55`):**
   Used `strconv.Itoa(i+1)` (loop index) instead of `len(args)` for `$N` placeholders. Now uses `len(args)` with a regression test in `model/summary_test.go`.
2. **[Fixed] UUIDv7 flagged as invalid (`model/fix/fix.go:439-440`):**
   `invalidUUIDs` previously checked `[1-5]` without a `$` anchor, flagging UUIDv7 records generated by `model/insert.go`. Updated to `[1-8]` with a `$` end anchor per RFC 9562.
3. **[Fixed] `deletedat IS NULL` in `HAVING` instead of `WHERE` (`internal/postgres/sql.go`):**
   Moved `files.deletedat IS NULL` into `WHERE` and removed `files.deletedat` from `GROUP BY` across all 11 releaser/BBS/FTP/magazine queries and `Role.Distinct()`.
4. **[Fixed] Two-query updates (`model/update.go`):**
   Replaced `OneFile` (`SELECT *`) + `f.Update(..., boil.Infer())` with a single targeted `boil.Whitelist(updatedat, <cols>...)` `UPDATE` via `updateCols`, returning `sql.ErrNoRows` if no row matched. Full read-back assertions added to `model/update_test.go`.
5. **[Fixed] `VACUUM ANALYZE` before `tx.Commit()` (`internal/config/fixer.go`, `model/fix/fix.go`):**
   Exported `fix.Optimize(ctx, db)`, removed the unused `db *sql.DB` parameter from `fix.Repair.Run`, and moved `fix.Optimize(ctx, db)` to run *after* `tx.Commit()` in `databaseFix`.
6. **[Fixed] `sql.DB` connection pool tuning (`internal/postgres/postgres.go`):**
   Configured `SetMaxOpenConns(25)`, `SetMaxIdleConns(10)`, `SetConnMaxLifetime(30m)`, and `SetConnMaxIdleTime(5m)` in `postgres.Open()`.
7. **[Fixed] Redundant statements in `SyncFilesIDSeq` (`model/fix/fix.go`):**
   Collapsed the 3-statement query into `SELECT setval('"files_id_seq"', COALESCE((SELECT MAX(id) FROM files), 1));` (eliminates discarded `SELECT`/`nextVal` calls, handles empty tables, and stops skipping an ID on each sync).
8. **[Fixed] `OR` chains replaced with `IN` (`model/artifacts.go`, `model/fix/fix.go`):**
   Refactored `onlyUniqueIDs` and `trainers` to use SQLBoiler's typed `.IN(...)` clauses instead of chained `id = ? OR id = ? ...`.
9. **[Fixed] Remote Demozoo/Pouet transactions & double-commit (`handler/htmx/submitprod.go`, `handler/routerlock.go`, `handler/app/remote/`):**
   Switched `DemozooLink`/`PouetLink` to `boil.ContextExecutor` and passed `db` directly instead of holding open transactions during remote HTTP downloads. Also removed a stray `tx.Commit()` in `PouetLink.Update` (double-commit bug) and ensured metadata saved on failed downloads is no longer rolled back.
10. **[Fixed] `SimilarTo*`, `ScenerSQL`, and `Sum*` SQL bugs (`internal/postgres/sql.go`, `model/releaser.go`):**
    Filtered before `GROUP BY`, replaced `SIMILAR TO` with escaped `LIKE`/`IN` + `deletedat IS NULL`, fixed the `ScenerSQL` middle-name pattern (`$2`), and added `COALESCE(..., 0)`, `deletedat IS NULL`, and `(group_brand_for = $1 OR group_brand_by = $1)` to `Sum*`.
11. **[Fixed] Broken `fixANSI` transaction, `EditFn` rollback guard, & mutex deadlock (`handler/app/dirs.go`, `handler/app/context.go`):**
    Removed the nil-`tx` crash in `fixANSI`, added `defer tx.Rollback()` in `EditFn`, and fixed a missing `mu.Unlock()` deadlock in `handler/app/context.go`.
12. **[Fixed] Duplicate-aware releaser suggestions & partial UUID search (`model/fix/suggest.go`, `handler/htmx/`, `model/artifacts.go`):**
    Added cached 3-tier fuzzy releaser suggestions for `<datalist>` (invalidated on edits/uploads) and `OnlyUniqueIDsWithPrefixes` using `lower(uuid) LIKE ? ESCAPE '\'` for `/editor/search/id`.
13. **[Fixed] Consolidated multi-query aggregations (`handler/internal/fileslice/stats.go`, `model/count.go`):**
    Collapsed `Stats.Get` from 23 concurrent queries into 1 `GROUP BY BTRIM(section), BTRIM(platform)` query, and `model.Counts` from 3 `errgroup` queries into 1 `COUNT(*) FILTER (WHERE ...)` query (`13.9 ms → 10.9 ms`, `8.4 KB → 1.3 KB`, `138 → 29` allocs).
14. **[Fixed] Programmatic PostgreSQL indexing (`model/fix/indexes.go`, `flags/flags.go`, `model/query_bench_test.go`):**
    Added the `index` (`i`) CLI subcommand (`defacto2-server index`) to build the pruned 11-index set concurrently, drop bloated legacy MySQL-migrated indexes, run `REINDEX TABLE CONCURRENTLY files`, and run `VACUUM ANALYZE files`.

### Remaining Optional Cleanups (No DB Upgrade Required)

- Collapse the ~55 repetitive category/platform filter structs in `model/filter.go` & `model/summary.go` (~2,800 lines) into a single table-driven helper.
- Convert padded `char(N)` columns (`platform`, `section`, `file_integrity_strong`) to `varchar`/`text` in Postgres so values are not right-padded with spaces.

---

## PostgreSQL Indexing & Benchmark Results

Indexes are managed programmatically by `fix.Indexes` (`model/fix/indexes.go`) via the CLI command:

```bash
defacto2-server index   # short alias: i
```

All `CREATE INDEX`, `DROP INDEX`, and `REINDEX TABLE` statements use `CONCURRENTLY` (zero read/write table locks), and `pg_trgm` is enabled with graceful degradation if unavailable.

### Active Index Set (66,112 rows, 29 MB heap)

Dropping the 2 legacy MySQL/pgloader indexes (`idx_16386_pubfor_pubby_pubedition_filename_comment` at 19 MB, which B-tree indexed the `text` column `comment`, and `idx_16386_browsing` at 15 MB) and pruning 2 redundant candidates (`idx_files_uuid` merged into `idx_files_uuid_lower_pattern`, and `idx_files_section_platform_date` dropped) reduced total index storage from **40 MB → 26 MB (35% smaller)**.

| Index | Definition | Size | Purpose |
| :--- | :--- | ---: | :--- |
| `idx_files_integrity` | `(file_integrity_strong) WHERE file_integrity_strong IS NOT NULL` | 7.98 MB | Upload SHA-384 duplicate check (`ExistHash`) |
| `idx_files_filename_trgm` | `GIN (filename gin_trgm_ops) WHERE deletedat IS NULL` | 5.73 MB | Substring `/search/file` (`ILIKE '%...%'`) |
| `idx_files_uuid_lower_pattern` | `(lower(uuid) text_pattern_ops)` | 3.84 MB | Exact `=` / `IN` and prefix `LIKE` UUID lookups |
| `idx_files_desc_fts` | `GIN (to_tsvector('english', COALESCE(record_title, '') \|\| ' ' \|\| COALESCE(comment, ''))) WHERE deletedat IS NULL` | 3.44 MB | Full-text `/search/desc` (`OnlyDescriptions`, `ByDescription`) |
| `idx_files_section_date` | `(section, date_issued_year, date_issued_month, date_issued_day) WHERE deletedat IS NULL` | 1.54 MB | Section category listings sorted by date |
| `idx_16386_primary` | `PRIMARY KEY (id)` | 1.46 MB | Primary key (shrunk from 7.1 MB bloated after `REINDEX`) |
| `idx_files_platform_date` | `(platform, date_issued_year, date_issued_month, date_issued_day) WHERE deletedat IS NULL` | 1.38 MB | Platform category listings sorted by date |
| `idx_files_upper_group_for` | `(upper(group_brand_for)) WHERE deletedat IS NULL` | 816 KB | Releaser pages (`/g/:name`) |
| `idx_files_demozoo` | `(web_id_demozoo) WHERE web_id_demozoo IS NOT NULL` | 344 KB | Demozoo ID lookups (`OneDemozoo`, `ExistDemozoo`) |
| `idx_files_pouet` | `(web_id_pouet) WHERE web_id_pouet IS NOT NULL` | 216 KB | Pouet ID lookups (`OnePouet`, `ExistPouet`) |
| `idx_files_upper_group_by` | `(upper(group_brand_by)) WHERE deletedat IS NULL AND group_brand_by IS NOT NULL` | 136 KB | Secondary releaser lookups |
| **Total (11 indexes)** | | **26 MB** | *(vs. **40 MB** across 3 legacy indexes)* |

### Read & Write Benchmarks

Measured against local PostgreSQL 16 (66,112 rows) with `go test -run=^$ -bench=^BenchmarkIndex -benchtime=500ms -count=3 ./model` (medians of 3 runs):

#### Reads

| Benchmark | Legacy (3 idx, 40 MB) | Initial (13 idx, 33 MB) | Pruned (11 idx, 26 MB) | Legacy → Pruned |
| :--- | ---: | ---: | ---: | ---: |
| `SearchDescriptions` | 442.8 ms | 3.25 ms | **2.95 ms** | **150× faster** |
| `UUIDPrefix` | 37.8 ms | 0.34 ms | **0.33 ms** | **115× faster** |
| `DedupeLookups` | 19.8 ms | 0.59 ms | **0.57 ms** | **35× faster** |
| `OneByUUID` | 5.46 ms | 0.22 ms | **0.21 ms** | **26× faster** |
| `SearchFilenames` | 167.1 ms | 45.4 ms | **47.3 ms** | **3.5× faster** |
| `ReleasersWhere` | 68.1 ms | 20.4 ms | **20.9 ms** | **3.3× faster** |
| `CategoryRare` (`Java`/`Console`/`Macos`/`TrialCrackme`) | 10.62 ms | 3.20 ms | **3.51 ms** | **3.0× faster** |
| `CategoryList` (`Nfo` + `IntroMsDos` page 1) | 1.98 ms | 2.30 ms | **2.18 ms** | **~1.0× (regression guard)** |

#### Writes (includes `BeginTx` + query + `Rollback`)

| Benchmark | Legacy (3 idx) | Initial (13 idx) | Pruned (11 idx) | 10 idx (no `trgm`) | Legacy → Pruned |
| :--- | ---: | ---: | ---: | ---: | ---: |
| `WriteInsert` (live row hitting all indexes) | 0.69 ms | 0.81 ms | **0.82 ms** | 0.79 ms | **+0.13 ms (+19%)** |
| `WriteUpdateHOT` (unindexed col via `boil.Whitelist`) | 0.48 ms | 0.46 ms | **0.48 ms** | 0.56 ms | **~0.00 ms (no regression)** |
| `WriteUpdateIndexed` (6 indexed cols incl. `comment`, `filename`) | 0.48 ms | 0.60 ms | **0.55 ms** | 0.54 ms | **+0.07 ms (+15%)** |

#### Write-Regression Notes

- **Unindexed edits stay on PostgreSQL's HOT (Heap-Only Tuple) path** with zero index writes, enabled by switching `model/update.go` from `boil.Infer()` to targeted `boil.Whitelist(...)` updates.
- **Pruning `idx_files_uuid` and `idx_files_section_platform_date`** saved 6.66 MB of index space and shaved ~0.05 ms off worst-case indexed `UPDATE`s (`0.60 ms → 0.55 ms`) with negligible read impact.
- **GIN `fastupdate` caveat:** Both GIN indexes (`idx_files_desc_fts`, `idx_files_filename_trgm`) buffer new entries in a pending list until `gin_pending_list_limit` (default 4 MB) or `autovacuum` flushes them into the main tree. Per-write overhead stays around +0.07–0.13 ms in steady state, with an occasional few-millisecond flush after large batches of edits/uploads — an easy trade-off for a read-dominant archive.

---

## Suggested Path Forward

- **Lowest risk / best ROI:** Run `defacto2-server index` on production, upgrade PostgreSQL to **18** (or **19** once comfortable with a fresh major release), and migrate from SQLBoiler to **Bob** or **`sqlc`**.
- **Maximum operational simplicity & preservation:** Migrate to **SQLite + `uptrace/bun`** (or **Bob** / **`sqlc`**) in a single coordinated refactor so the query layer is only rewritten once.
