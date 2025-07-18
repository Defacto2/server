# Database readme

Previous iterations of the Defacto2 web application relied on MySQL for its database. But for this 2023, Go application rewrite, the site will use [PostgreSQL](https://www.postgresql.org). With the Defacto2 database schema, Postgre uses less memory and is far more performant with complex queries.

Postgre is more strict about data types than MySQL. For example, inserting a string into a numeric column in Postgre will throw an error, whereas MySQL will convert the string to a number. Postgre has a more powerful query optimizer meaning queries often run faster with complex joins and subqueries.

## Troubleshoot

Troubleshoot after syncing the database with an backed up sql file.

> f.Insert: models: unable to insert into files: ERROR: duplicate key value violates unique constraint "idx_16386_primary" (SQLSTATE 23505) 5454

To fix on Docker, run the following command:

The Docker container name is `postgres16` and the database name is `defacto2_ps`.
The default username is `root` and the password is `example`.

```sh
docker exec -it postgres16 psql --username=root --dbname=defacto2_ps

> \connect defacto2_ps
```

In the PostgreSQL shell, run the following SQL commands:

```sql
SELECT nextval('files_id_seq'), max(id) FROM files;

 nextval |  max
---------+-------
   52779 | 55297
(1 row)
```

The nextval is **less** than the max value, so we need to reset the sequence.

```sql
SELECT nextval('files_id_seq'), max(id) FROM files;

 setval
--------
  55297
(1 row)
```

Check the sequence value again to ensure it has been reset.

```sql
SELECT nextval('files_id_seq'), max(id) FROM files;

 nextval |  max
---------+-------
   52779 | 55297
(1 row)

```

# Table and data <small>changes to implement</small>

These are only suggestions and may not be necessary if they create too much work or complexity.

- [ ] Rename `files` table to `release` or `releases`
- [ ] Create a `release_tests` table with a selection of 20 read-only records
- [ ] Rename `files.createdat`, `deleteat`, `updatedat` etc to `[name]_at` aka `create_at`...
- - [ ] **OR** break convention and use `date_created`, `date_deleted`, `date_updated` etc.
- [ ] DROP `dosee_no_aspect_ratio_fix`.
- - > `ALTER TABLE "files" DROP "dosee_no_aspect_ratio_fix"; COMMENT ON TABLE "files" IS '';`

### [datatypes](https://www.postgresql.org/docs/current/datatype.html) differences

- [ ] `CITEXT` type for case-insensitive character strings
- [ ] `files.filesize` should be converted to an `integer`, 4 bytes to permit a 2.147GB value
- [ ] `files.id` should be converted to a `serial` type
- [ ] There is no performance improvement for `fixed-length`, padded character types, etc, meaning strings can use `varchar`(n) or `text`.

### Indexes

- [ ] Create PostgreSQL _indexes_ with case-sensitive strings for [optimal performance](https://wirekat.com/optimizing-sql-based-on-postgresql/)?
- [ ] Partial Indexes: Use partial indexes when you only need to index a subset of rows, such as,
- - `CREATE INDEX ON orders (order_date) WHERE status = 'SHIPPED'`;
- [ ] Over-Indexing: Creating too many indexes can slow down write operations, as each index needs to be updated on `INSERT`, `UPDATE`, or `DELETE` operations.
- [ ] Index Maintenance: Rebuild indexes periodically to deal with bloat using `REINDEX`.
- [ ] Indexing Join Columns: Index columns that are used in JOIN conditions to improve join performance.
  > `combineGroup` and `(r Role) Distinct()`

### Future idea, _file archive content_ relationship table

Create a relationship files table that contains the filename content within of every archive release.

We could also include columns containing size in bytes, sha256 hash, text body for full text searching.

This would replace the `file_zip_content` column and also, create a CLI tool to scan the archives to fill out this data. For safety and code maintenance, the tool would need to be a separate program from the web server application.

## Migration from MySQL to PostgreSQL

This document describes how to migrate the Defacto2 MySQL database to PostgreSQL using [pgloader](https://pgloader.io/). Note, the migration is a one-time operation and should be run on a development or staging server before running on the production server.

- `defacto2-inno` is the name of the MySQL database.
- `defacto2_ps` is the name of the PostgreSQL database, note the `_` in the name as opposed to `-`.

Create a migration loader file named `migrate.load` with the following content, replacing the connection strings with your own database credentials:

```sql
LOAD DATABASE
     FROM     mysql://root:example@localhost:3306/defacto2-inno?useSSL=false
     INTO     pgsql://root:example@localhost:5432/defacto2_ps

 WITH include drop, create tables

ALTER SCHEMA 'source_db' RENAME TO 'public'
;
```

Run the migration using the following command:

```sh
# run the migration
pgloader migrate.load

# test the migration
postgres psql
```

```sql
# SELECT * FROM files;
```

A simple client application to interact with the migrated database is [Postbird](https://github.com/paxa/postbird).

Some more resources:

- [pgloader](https://pgloader.io/)
- [documentation](https://pgloader.readthedocs.io/en/latest/)
- [Docker hub](https://hub.docker.com/r/dimitri/pgloader/)
- [DigitalOcean how-to migrate](https://www.digitalocean.com/community/tutorials/how-to-migrate-mysql-database-to-postgres-using-pgloader)
- [Migrating from MySQL to PostgreSQL](https://pgloader.readthedocs.io/en/latest/tutorial/tutorial.html#migrating-from-mysql-to-postgresql)
