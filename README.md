# Defacto2 server

Under construction.

#### TODOs

Mock the database using a test sqlite db?
Boil eager loading of groups etc. known as the n+1 problem.
 - It takes a loop of queries, and merges them into a single big statement for db performance

Rename Files database to Releases / Release.

# PostgreSQL migrations
# https://www.postgresql.org/docs/current/datatype.html
# CITEXT = case-insensitive character string type.
# size byes convert to `integer` (4 bytes) = max 2.147GB value
# id = serial

rename created_at updated_at

There is no performance improvement for fixed-length, padded character types.
So should always use varchar(n) or text

uuid type = may need conversion to legacy cfml? // a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11
// could create a func to rename all legacy filenames on the host system to use rfc 4122
// 8-4-4-4-12 to equal 32 digits

date_issued_month etc should have range limits 1-12

todo: store NFOs,file_id.diz in database using the bytea hex format data type?
hex is more performant than binary escape format
https://www.postgresql.org/docs/current/datatype-binary.html

Full text search types.
https://www.postgresql.org/docs/current/datatype-textsearch.html

todo: create a relationship files table that contains the filename of every
file contained within a zip archive. could also include columns containing size in bytes, sha256 hash, text body for full text searching. this woul replace the file_zip_content column
would also create a sep cli tool to scan the archives to fill out this data.
for saftey and code maintainance maintenance, the tool needs to be sep from `server` and `df2`

All SQL needs to account for delete_at
`qm.WithDeleted`

Create a new postgresql files db for testing.
	defacto2-tests > files, groupnames
    sudo -u postgres -- createuser --createdb --pwprompt start
        createdb --username start --password --owner start gettingstarted
    psql --username start --password gettingstarted < schema.sql

```go
var users []*models.User

rows, err := db.QueryContext(context.Background(), `select * from users;`)
if err != nil {
    log.Fatal(err)
}
err = queries.Bind(rows, &users)
if err != nil {
    log.Fatal(err)
}
rows.Close()
```

# see: https://pgloader.readthedocs.io
# quick guide: https://thedevelopercafe.com/articles/sql-in-go-with-sqlboiler-ac8efc4c5cb8

```go
# https://github.com/volatiletech/sqlboiler#extending-generated-models
# extending generated models
# method 1: simple funcs

// Package modext is for SQLBoiler helper methods
package modext

// UserFirstTimeSetup is an extension of the user model.
func UserFirstTimeSetup(ctx context.Context, db *sql.DB, u *models.User) error { ... }

# calling code

user, err := Users().One(ctx, db)
// elided error check

err = modext.UserFirstTimeSetup(ctx, db, user)
// elided error check


```

#### Later TODOs

- Create a method to calc the most popular years for a record query.
- Optimise the columns fetched from SQL queries, using bindings.
- OrderBy Name/Count /html3/groups? https://pkg.go.dev/sort#example-package-SortKeys
- Add both group_for and group_by to distinct results.
- Move orderby params to cookies?

### echo tools to implement

- SQLBoiler binding https://thedevelopercafe.com/articles/sql-in-go-with-sqlboiler-ac8efc4c5cb8

- Request and validate data (forms, query, etc) https://echo.labstack.com/guide/request/#custom-binder
- Response for future APIs https://echo.labstack.com/guide/response/
- Before / After responces to handle data outside of templates https://echo.labstack.com/guide/response/#hooks
- Custom context that are used in handlers https://echo.labstack.com/guide/context/
- HTTPS Server custom context that are used in handlers - https://echo.labstack.com/guide/context/
     https://echo.labstack.com/cookbook/http2/
- HTTP/2 Cleartext Server
- CORS https://echo.labstack.com/middleware/cors/
     https://echo.labstack.com/cookbook/cors/
- CSRF Middleware https://echo.labstack.com/middleware/csrf/
- Route tests! https://echo.labstack.com/guide/testing/
- List all routes for testing and debugging https://echo.labstack.com/guide/routing/#list-routes
- Testing middleware examples https://github.com/labstack/echo/tree/master/middleware
- Body dump middleware when in dev mode https://echo.labstack.com/middleware/body-dump/
- Body limit middleware except for uploads? https://echo.labstack.com/middleware/body-limit/
- Rate Limiter Middleware (needed for 16k concurrent requests) https://echo.labstack.com/middleware/rate-limiter/
- Request ID Middleware https://echo.labstack.com/middleware/request-id/
- Secure Middleware https://echo.labstack.com/middleware/secure/
- Session middleware https://echo.labstack.com/middleware/session/
- Timeout middleware for specific routes https://echo.labstack.com/middleware/timeout/
- Streaming response https://echo.labstack.com/cookbook/streaming-response/


- User accounts with CRUD https://echo.labstack.com/cookbook/crud/
- File uploads https://echo.labstack.com/cookbook/file-upload/
     https://echo.labstack.com/cookbook/file-upload/
     https://echo.labstack.com/cookbook/timeouts/

- Sub Domains https://echo.labstack.com/cookbook/subdomains/

#### Authentication options

Basic auth middleware provides an HTTP basic authentication.

    For valid credentials it calls the next handler.
    For missing or invalid credentials, it sends “401 - Unauthorized” response.

https://echo.labstack.com/middleware/basic-auth/

Casbin Auth Middleware

    Dependencies
    Custom Configuration
    Configuration

Casbin is a powerful and efficient open-source access control library for Go. It provides support for enforcing authorization based on various models. 

https://echo.labstack.com/middleware/casbin-auth/
https://github.com/casbin/casbin
https://casbin.org/

JWT Middleware

    Custom Configuration
    Configuration
    Example

JWT provides a JSON Web Token (JWT) authentication middleware. Echo JWT middleware is located at https://github.com/labstack/echo-jwt
https://echo.labstack.com/middleware/jwt/
https://echo.labstack.com/cookbook/jwt/


Key Auth Middleware

    Custom Configuration
    Configuration

Key auth middleware provides a key based authentication.

    For valid key it calls the next handler.
    For invalid key, it sends “401 - Unauthorized” response.
    For missing key, it sends “400 - Bad Request” response.

https://echo.labstack.com/middleware/key-auth/

#### IP config options

https://echo.labstack.com/guide/ip-address/

Case 1. With no proxy

If you put no proxy (e.g.: directory facing to the internet), all you need to (and have to) see is IP address from network layer. Any HTTP header is untrustable because the clients have full control what headers to be set.

In this case, use echo.ExtractIPDirect().

Case 2. With proxies using X-Forwarded-For header

X-Forwared-For (XFF) is the popular header to relay clients’ IP addresses. At each hop on the proxies, they append the request IP address at the end of the header.

Case 3. With proxies using X-Real-IP header

X-Real-IP is another HTTP header to relay clients’ IP addresses, but it carries only one address unlike XFF.

#### Binding request data

Data Sources

Echo supports the following tags specifying data sources:

    query - query parameter
    param - path parameter (also called route)
    header - header parameter
    json - request body. Uses builtin Go json package for unmarshalling.
    xml - request body. Uses builtin Go xml package for unmarshalling.
    form - form data. Values are taken from query and request body. Uses Go standard library form parsing.

Data Types

When decoding the request body, the following data types are supported as specified by the Content-Type header:

    application/json
    application/xml
    application/x-www-form-urlencoded

When binding path parameter, query parameter, header, or form data, tags must be explicitly set on each struct field. However, JSON and XML binding is done on the struct field name if the tag is omitted. This is according to the behaviour of Go’s json package.

---

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
FAIL	github.com/Defacto2/server/models	0.010s
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

#### OPENAPI?

https://github.com/deepmap/oapi-codegen
https://threedots.tech/post/serverless-cloud-run-firebase-modern-go-application/#public-http-api

#### Event / Message applications

Go library for building event-driven applications.

https://watermill.io/
https://github.com/ThreeDotsLabs/watermill
https://github.com/ThreeDotsLabs/watermill/tree/master/_examples

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