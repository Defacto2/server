# TODOs and tasks

### Recommendations

- [ ] Create only one DB connection sitewide, and use .Ping() to test the connection on startup.
- [ ] 
- [ ] Use DigitalOcean API to display Estimated Droplet Transfer Pool usage and remaining balance. 
		https://pkg.go.dev/github.com/digitalocean/godo https://docs.digitalocean.com/reference/api/api-reference/#operation/droplets_get
- [ ] After a successful demozoo/pouet upload, defer a sync for the data to the artifact record.
- [ ] 
- [ ] Missing screenshots and downloads: https://go.defacto2.net/g/millennium-ftp
- [ ] 
- [ ] On startup, modify downloads to use database stored, last modified value.
- [ ] 
- [ ] On startup, run magic numbers on all records to replace the current value in database.
- [ ] 
- [ ] Complete `internal/archive/archive.go to support all archive types. need to supprt legacy zip via hwzip and arc.
- [ ] 
- [ ] Render HTML in an iframe instead of readme? Example, http://localhost:1323/f/ad3075
- [ ] 
- [ ] Handle magazines title on artifact page, http://localhost:1323/f/a55ed, this is hard to read, "Issue 4\nThe Pirate Syndicate +\nThe Pirate World"
- [ ] 
- [ ] If artifact is a text file displayed in readme, then delete image preview, these are often super long, large and not needed.
- [ ] 
- [ ] If a #hash is appended to a /f/<id> URL while signed out, then return a 404 or a redirect to the sign in page. Post signing should return to the #hash URL?
- [ ] 
- [ ] Delete all previews that are unused, such as textfiles that are displayed as a readme.
- [ ] 
- [ ] On Demozoo or Pouet upload or reach, locally cache the JSON to the temp directory.
- [ ] 
- [ ] Fix, file editor menu close links not working when #hash is appended to the URL.

- [ ] - http://www.platohistory.org/
- [ ] - https://portcommodore.com/dokuwiki/doku.php?id=larry:comp:bbs:about_cbbs
- [ ] - 8BBS https://everything2.com/title/8BBS



Magic files to add:

- Excel, http://localhost:1323/f/b02fcc
- Multipart zip, http://localhost:1323/f/a9247de, http://localhost:1323/f/a619619
- Convert ms-dos filenames to valid utf-8 names, see http://localhost:1323/f/b323bda
- Windows binaries

### Locations

#### On startup tasks

 - `server.go` 
 - - checks()
 - - repairs() ~ 
 - - repairDatabase() ~ /model/fix/fix.go ~ Artifacts.Run()

### Bug fixes

### Emulate on startup fixes

- [ ] Repack zips that contain programs with bad filenames, for example: http://localhost:1323/f/ab252e4

### Templates

- [ ] `artifactfile.tmpl`
- [ ] `layoutjs.tmpl`
 
---

#### Upload tests

24 June.

- [X] Demozoo Prod.
- [X] Demozoo Graphic.
- [X] Intros.
- [X] Trainer.
- [X] Installer.
- [X] PC and Amiga text.
- [X] Image brand, logo or proof.
- [X] Text, DOS and Windows magazines.
- [X] Advanced.

#### Software libraries

#####  Scheduling library for Go

- [gocron](https://github.com/go-co-op/gocron)
- [Go Quartz](https://github.com/reugn/go-quartz)


### Error PANIC

```go
10:27:59	ERROR	app/error.go:166	500 internal error for the URL, "artifact": write tcp 127.0.0.1:1323->127.0.0.1:39474: write: broken pipe: caused by artifact b9442d (8887)
github.com/Defacto2/server/handler/app.InternalErr
	/home/ben/github/server/handler/app/error.go:166
github.com/Defacto2/server/handler/app.Dirs.Artifact
	/home/ben/github/server/handler/app/dirs.go:192
github.com/Defacto2/server/handler.Configuration.website.func3
	/home/ben/github/server/handler/router.go:208
github.com/labstack/echo/v4.(*Echo).add.func1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:587
github.com/Defacto2/server/handler.Configuration.nonce.Middleware.MiddlewareWithConfig.func1.1
	/home/ben/go/pkg/mod/github.com/labstack/echo-contrib@v0.17.1/session/session.go:73
github.com/Defacto2/server/handler.Configuration.Controller.RemoveTrailingSlashWithConfig.func2.1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/slash.go:117
github.com/labstack/echo/v4/middleware.RequestLoggerConfig.ToMiddleware.func1.1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/request_logger.go:286
github.com/Defacto2/server/handler.Configuration.Controller.Secure.SecureWithConfig.func4.1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/secure.go:141
github.com/labstack/echo/v4.(*Echo).ServeHTTP.func1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:668
github.com/Defacto2/server/handler.Configuration.Controller.NonWWWRedirect.NonWWWRedirectWithConfig.redirect.func6.1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/redirect.go:152
github.com/labstack/echo/v4/middleware.RewriteWithConfig.func1.1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/rewrite.go:77
github.com/labstack/echo/v4.(*Echo).ServeHTTP
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:674
net/http.serverHandler.ServeHTTP
	/usr/local/go/src/net/http/server.go:3142
net/http.(*conn).serve
	/usr/local/go/src/net/http/server.go:2044
10:27:59	ERROR	app/error.go:183	500 internal render error for the URL, "artifact": the server could not render the html template for this page
github.com/Defacto2/server/handler/app.InternalErr
	/home/ben/github/server/handler/app/error.go:183
github.com/Defacto2/server/handler/app.Dirs.Artifact
	/home/ben/github/server/handler/app/dirs.go:192
github.com/Defacto2/server/handler.Configuration.website.func3
	/home/ben/github/server/handler/router.go:208
github.com/labstack/echo/v4.(*Echo).add.func1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:587
github.com/Defacto2/server/handler.Configuration.nonce.Middleware.MiddlewareWithConfig.func1.1
	/home/ben/go/pkg/mod/github.com/labstack/echo-contrib@v0.17.1/session/session.go:73
github.com/Defacto2/server/handler.Configuration.Controller.RemoveTrailingSlashWithConfig.func2.1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/slash.go:117
github.com/labstack/echo/v4/middleware.RequestLoggerConfig.ToMiddleware.func1.1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/request_logger.go:286
github.com/Defacto2/server/handler.Configuration.Controller.Secure.SecureWithConfig.func4.1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/secure.go:141
github.com/labstack/echo/v4.(*Echo).ServeHTTP.func1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:668
github.com/Defacto2/server/handler.Configuration.Controller.NonWWWRedirect.NonWWWRedirectWithConfig.redirect.func6.1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/redirect.go:152
github.com/labstack/echo/v4/middleware.RewriteWithConfig.func1.1
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/rewrite.go:77
github.com/labstack/echo/v4.(*Echo).ServeHTTP
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:674
net/http.serverHandler.ServeHTTP
	/usr/local/go/src/net/http/server.go:3142
net/http.(*conn).serve
	/usr/local/go/src/net/http/server.go:2044
10:27:59	DPANIC	config/error.go:54	Custom response handler broke: %swrite tcp 127.0.0.1:1323->127.0.0.1:39474: write: broken pipe
github.com/Defacto2/server/internal/config.Config.CustomErrorHandler
	/home/ben/github/server/internal/config/error.go:54
github.com/labstack/echo/v4.(*Echo).ServeHTTP
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:675
net/http.serverHandler.ServeHTTP
	/usr/local/go/src/net/http/server.go:3142
net/http.(*conn).serve
	/usr/local/go/src/net/http/server.go:2044
```