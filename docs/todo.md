# TODOs and tasks

### Ideas or doubleful tasks

- [ ] Implememnt a scheduling library for Go
- - [gocron](https://github.com/go-co-op/gocron)
- - [Go Quartz](https://github.com/reugn/go-quartz)

### Recommendations

- [ ] _Unique ID_ and _Record Key_ buttons should copy the value to the clipboard.
- [ ] Use DigitalOcean API to display Estimated Droplet Transfer Pool usage and remaining balance. 
		https://pkg.go.dev/github.com/digitalocean/godo https://docs.digitalocean.com/reference/api/api-reference/#operation/droplets_get
- [ ] js-dos doesn't yet support `extras` zipped files.
- [ ] After a successful demozoo/pouet upload, defer a sync for the data to the artifact record.
- [ ] Find all uuid zip files that use legacy zip encodings, EXPLODE, UNSHRINK, UNREDUCE.
- [ ] Find all msdos apps that use incompatible archives such as LHA, ARJ, ARC.

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

### Bug fixes

- [ ] htmx search for id or uuid, incomplete (less than 32 chars) uuid should not trigger a search.

---

### Errors when displaying artifacts

```
10:56:03	ERROR	app/error.go:165	500 error for "artifact": write tcp 127.0.0.1:1323->127.0.0.1:52014: write: broken pipe
github.com/Defacto2/server/handler/app.InternalErr
	/home/ben/github/server/handler/app/error.go:165
github.com/Defacto2/server/handler/app.Dirs.Artifact
	/home/ben/github/server/handler/app/dirs.go:155
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
	/usr/local/go/src/net/http/server.go:3137
net/http.(*conn).serve
	/usr/local/go/src/net/http/server.go:2039
10:56:03	ERROR	app/error.go:182	the server could not render the html template for this page: write tcp 127.0.0.1:1323->127.0.0.1:52014: write: broken pipe
github.com/Defacto2/server/handler/app.InternalErr
	/home/ben/github/server/handler/app/error.go:182
github.com/Defacto2/server/handler/app.Dirs.Artifact
	/home/ben/github/server/handler/app/dirs.go:155
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
	/usr/local/go/src/net/http/server.go:3137
net/http.(*conn).serve
	/usr/local/go/src/net/http/server.go:2039
10:56:03	DPANIC	config/error.go:54	Custom response handler broke: %swrite tcp 127.0.0.1:1323->127.0.0.1:52014: write: broken pipe
github.com/Defacto2/server/internal/config.Config.CustomErrorHandler
	/home/ben/github/server/internal/config/error.go:54
github.com/labstack/echo/v4.(*Echo).ServeHTTP
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:675
net/http.serverHandler.ServeHTTP
	/usr/local/go/src/net/http/server.go:3137
net/http.(*conn).serve
	/usr/local/go/src/net/http/server.go:2039
```