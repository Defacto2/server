# Database readme

Previous iterations of the Defacto2 web application relied on MySQL for its database. But for this 2023 application rewrite, the site will use PostgreSQL. 

Postgre is more strict about data types than MySQL. For example, inserting a string into a numeric column in Postgre will throw an error, whereas MySQL will convert the string to a number. Postgre has a more powerful query optimizer meaning queries often run faster with complex joins and subqueries.

## Table and data changes to implement

- Rename `Files` table to `Release` or `Releases`
- Create a `Release_tests` table with a selection of 20 read-only records.
- Rename `files.createdat` etc to `?_at` aka `create_at`.

[PostgreSQL datatypes](https://www.postgresql.org/docs/current/datatype.html)

`CITEXT` type for case-insensitive character strings.

`files.filesize` should be converted to an `integer`, 4 bytes to permit a 2.147GB value.

`files.id` should be converted to a `serial` type.

### There is no performance improvement for `fixed-length`, padded character types, 
Meaning strings can use `varchar`(n) or `text`.

#### UUID

`files.UUID` have be renamed from CFML style to the universal RFC-4122 syntax.

This will require the modification of queries when dealing with `/files/[uuid|000|400]`.

CFML is 35 characters, 8-4-4-16.
`xxxxxxxx-xxxx-xxxx-xxxxxxxxxxxxxxxx`

RFC is 36 characters, 8-4-4-4-12.
`xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`

#### Store NFOs and texts

We can store NFO and textfiles plus file_id.diz in database using the `bytea` hex-format, binary data type. It is more performant than the binary escape format.

https://www.postgresql.org/docs/current/datatype-binary.html

#### Full text seach types

https://www.postgresql.org/docs/current/datatype-textsearch.html

#### Files content relationship table

Create a relationship files table that contains the filename content within of every archive release. 

We could also include columns containing size in bytes, sha256 hash, text body for full text searching. 

This would replace the `file_zip_content` column and also, create a CLI tool to scan the archives to fill out this data. For saftey and code maintenance, the tool would need to be a separate program from the web server application.

## Migrate MySQL to PostgreSQL in a single, Docker command

[This command relies on pgloader](https://pgloader.io/).

> pgloader loads data into PostgreSQL and allows you to implement Continuous Migration from your current database to PostgreSQL. 

- `defacto2-inno` is the name of the MySQL database.
- `defacto2-ps` is the name of the PostgreSQL database.

```sh
docker run --network host --rm -it dimitri/pgloader:latest \
     pgloader --verbose \
       mysql://root:example@127.0.0.1/defacto2-inno \
       pgsql://root:example@127.0.0.1/defacto2-ps
```

- [pgloader](https://pgloader.io/)
- [documentation](https://pgloader.readthedocs.io/en/latest/)
- [Docker hub](https://hub.docker.com/r/dimitri/pgloader/)
- [DigitalOcean how-to migrate](https://www.digitalocean.com/community/tutorials/how-to-migrate-mysql-database-to-postgres-using-pgloader)

### Customize the migration

A psloader migration is customized with a configuration file which might be more stable than using the commandline.

[Migrating from MySQL to PostgreSQL](https://pgloader.readthedocs.io/en/latest/tutorial/tutorial.html#migrating-from-mysql-to-postgresql)

`inno-to-ps.load`
```sql
LOAD DATABASE
     FROM      mysql://pgloader_my:mysql_password@mysql_server_ip/source_db?useSSL=true
     INTO     pgsql://pgloader_pg:postgresql_password@localhost/new_db

 WITH include drop, create tables

ALTER SCHEMA 'source_db' RENAME TO 'public'
;
```

```sh
# run the migration
pgloader inno-to-ps.load
# test the migration
postgres psql
```

```sql
# SELECT * FROM files;
```
