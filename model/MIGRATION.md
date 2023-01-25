# Migration of databases README

## Migrate to PostgreSQL in a single command!

```sh
docker run --network host --rm -it dimitri/pgloader:latest \
     pgloader --verbose \
       mysql://root:example@127.0.0.1/defacto2-inno \
       pgsql://root:example@127.0.0.1/defacto2-ps
```

https://pgloader.io/

https://pgloader.readthedocs.io/en/latest/

https://hub.docker.com/r/dimitri/pgloader/

https://www.digitalocean.com/community/tutorials/how-to-migrate-mysql-database-to-postgres-using-pgloader

### TODO

Create a `mysql-to-ps.load` migration config

```sql
LOAD DATABASE
     FROM      mysql://pgloader_my:mysql_password@mysql_server_ip/source_db?useSSL=true
     INTO     pgsql://pgloader_pg:postgresql_password@localhost/new_db

 WITH include drop, create tables

ALTER SCHEMA 'source_db' RENAME TO 'public'
;
```

---

### Install the PostgreSQL client onto the host, Ubuntu system

Note, the client kept in the Ubuntu software repo is out of date.

Use these instructions to install the active version: 
https://www.postgresql.org/download/linux/ubuntu/

postgresql-client: `/usr/bin/dropdb`