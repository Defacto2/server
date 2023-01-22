# Defacto2 server

Under construction.

#### TODOs

Mock the database using a test sqlite db?

#### Later TODOs

- Create a method to calc the most popular years for a record query.
- Optimise the columns fetched from SQL queries, using bindings.
- OrderBy Name/Count /html3/groups? https://pkg.go.dev/sort#example-package-SortKeys
- Add both group_for and group_by to distinct results.
- Move orderby params to cookies?

### echo tools to implement

- SQLBoiler binding https://thedevelopercafe.com/articles/sql-in-go-with-sqlboiler-ac8efc4c5cb8

- Echo#*Server#ReadTimeout 
- Echo#*Server#WriteTimeout
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