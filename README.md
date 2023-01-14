
####

These cannot be mixed, otherwise mysqldump errors will occur!

```sh
sudo apt install mysql-client

or 

sudo apt install mariadb-client
```

mariadb causes issues with SQLBoiler :/
https://github.com/volatiletech/sqlboiler/issues/329


```
go test ./models      
failed running: mysql [--defaults-file=/tmp/optionfile3646082045]

mysql: unknown variable 'ssl-mode=DISABLED'

Unable to execute setup: exit status 7
FAIL	github.com/bengarrett/df2023/models	0.010s
FAIL
```

https://hub.docker.com/r/dimitri/pgloader/

docker run --network host --rm -it dimitri/pgloader:latest \
     pgloader --verbose \
       mysql://root:example@127.0.0.1/defacto2-inno \
       pgsql://root:example@127.0.0.1/defacto2-ps

todo: create a `mysql-to-ps.load` migration config
https://pgloader.readthedocs.io/en/latest/ref/mysql.html?highlight=schema#using-default-settings

todo: rename psql schema to public

#### mysql-to-ps.load

```
LOAD DATABASE
     FROM      mysql://pgloader_my:mysql_password@mysql_server_ip/source_db?useSSL=true
     INTO     pgsql://pgloader_pg:postgresql_password@localhost/new_db

 WITH include drop, create tables

ALTER SCHEMA 'source_db' RENAME TO 'public'
;
```
https://www.digitalocean.com/community/tutorials/how-to-migrate-mysql-database-to-postgres-using-pgloader

(14)
postgresql-client: /usr/bin/dropdb

Or use these instructions to install 15+
https://www.postgresql.org/download/linux/ubuntu/

Boiler guide: https://blog.logrocket.com/introduction-sqlboiler-go-framework-orms/

Live reloading
go install github.com/cosmtrek/air@latest
https://thedevelopercafe.com/articles/live-reload-in-go-with-air-4eff64b7a642

SQL in Go with SQLBoiler
https://thedevelopercafe.com/articles/sql-in-go-with-sqlboiler-ac8efc4c5cb8