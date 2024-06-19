# TODOs and tasks

  * (star) __*__ prefix indicates a *low priority* task.
  * (question) __?__ prefix indicates an *idea* or *doubtful*.

### Terminal commands and flags

- [ ] *? Command to clean up the database and remove all orphaned records.
- [ ] *? Command to reindex the database, both to erase and rebuild the indexes.

### Files and assets

- [ ] Create a htmx, live _classifications page_ for editors, using the advanced uploader `<select>` fields.
- [ ] ~~Data editor button should reload the page when data Editor module is active~~. Could not implement.
- [X] Fix PNG binary being displayed in the text viewer. http://localhost:1323/f/af20fcb,http://localhost:1323/f/a73208,http://localhost:1323/f/b34c2a,
- [-] Fix broken Unicode multibyte text being displayed in the text viewer. http://localhost:1323/f/b12d05f,http://localhost:1323/f/b228a8e,
- [ ] Create a new assets path named `extras` and move into it all the uuid files from `downloads` that use file extensions.

### Menus and layout

- [X] Create a menu link to DigitalOcean referal page, [or/and] add a link to the thanks page.
- [ ] Create a locked menu option to search the database by file ID or UUID or ~~URL~~.
- [ ] Create a locked page with links for various file items that use unique website features,
      for example, DOS emulation for different archives file types and raw executables.
      And also different text file types for the text viewer.
- [ ] Permenantly delete record should have its own modal dialog, requiring a confirmation and a reason using htmx result. Afterwards, a success message should be displayed and a notice that the active browser tab will now return a 404. Notes on reasons, only if an upload is unwanted, otherwise mention how deletion is permanent and breaks SEO and third-party links.
- [X] Use DigitalOcean API to display Estimated Droplet Transfer Pool usage and remaining balance.

### Database

- [ ] Create a DB fix to detect and rebadge msdos and windows trainers.
- [ ] `OrderBy` Name/Count /html3/groups? https://pkg.go.dev/sort#example-package-SortKeys

### Backend

- [ ] *? Implememnt a [scheduling library for Go](https://github.com/reugn/go-quartz)
- [ ] [xstrings](https://github.com/huandu/xstrings) for string manipulation.
- [ ] Errors cleanup, never return raw errors, always wrap them. And also never use, "xxx failed or broke" as an error message. Instead use "doing xxx" or "while doing xxx".


#### Future locked file items list for testing features.

- ~~Unknown codepage~~, prob cp437 accidently saved as unicode: http://localhost:1323/f/ac2319e,http://localhost:1323/f/b0269ca [comparison:http://localhost:1323/f/ac1d9d3],
- ~~GIF image~~: http://localhost:1323/f/b828636,http://localhost:1323/f/b42e22b,http://localhost:1323/f/ae2a407,
- ~~Excess tail~~ whitespace: http://localhost:1323/f/b830654,
- ~~Missing newlines~~, requires wrap: http://localhost:1323/f/b14bb1,http://localhost:1323/f/b12fe37,http://localhost:1323/f/ad23d9c,http://localhost:1323/f/b122787,
- ~~ [requires textfile extraction from zip] No text preview: http://localhost:1323/f/af31a9,
- ~~ HTML file preview: http://localhost:1323/f/a722b1f,
- ~~ PDF file preview: http://localhost:1323/f/b04139,
- ~~ Block text file: http://localhost:1323/f/ad217af,http://localhost:1323/f/ae2a9cc,http://localhost:1323/f/ad2b193,http://localhost:1323/f/ac1a2dc,
- ~~ [REQUIRES FIX] Multibyte Unicode [Unicode support is done through filemagic column]: http://localhost:1323/f/b12d05f,http://localhost:1323/f/b53028e,
- ~~IRL Link to~~: http://localhost:1323/f/b029330,http://localhost:1323/f/ba4805,http://localhost:1323/f/ab27f81,http://localhost:1323/f/b029330,
- ~~href in text viewer~~: http://localhost:1323/f/a92c1dc,http://localhost:1323/f/a734e9,http://localhost:1323/f/ac2a79,
- [REQUIRES FIX] JSDOS unsupported zip archive: http://localhost:1323/f/a22af8,http://localhost:1323/f/b72d52b,http://localhost:1323/f/b52e058,http://localhost:1323/f/b3286a2
- ~~CP437 text pattern detection~~: http://localhost:1323/f/ab2f2b4,http://localhost:1323/f/ba2db9e,http://localhost:1323/f/ab43cb,http://localhost:1323/f/b42ccd8,
- ~~Unicode single byte~~: http://localhost:1323/f/a5191c3,http://localhost:1323/v/ab1fc8b,http://localhost:1323/f/b61f24f,
- ~~Text viewer attempting~~ to preview PNG image due to category: http://localhost:1323/f/af20fcb,
- Maximum download permitted, 1GB: http://localhost:1323/f/aa256f1,
- Emulator, .exe file: http://localhost:1323/f/b23b9c,http://localhost:1323/f/ad18cf0,http://localhost:1323/f/b030891,http://localhost:1323/f/b33057
- Emulator, unsupported .arj: http://localhost:1323/f/ad18c43,http://localhost:1323/f/ae1c87b,http://localhost:1323/f/b41eefb,
- Emulator, unsupported .arc: http://localhost:1323/f/a418d56,
- Emulator, unsupported .lha: http://localhost:1323/f/a724be8,
- ~~Amiga ansi colors~~: http://localhost:1323/f/b02ecf1,
- ~~Text viewer~~, extra big textfiles, 5MB: http://localhost:1323/f/ae219e8,
- ~~Text viewer~~ UTF-16: http://localhost:1323/f/b22e849,

### MAJOR FREEZE

```
goroutine 18685 [running]:
github.com/h2non/filetype.Match({0xc0022ea000, 0x200, 0x200})
	/home/ben/go/pkg/mod/github.com/h2non/filetype@v1.1.3/match.go:28 +0x185
github.com/Defacto2/server/handler/app.artifactMagic({0xc001256000?, 0xc000ddb500?})
	/home/ben/github/server/handler/app/dirs.go:584 +0x1130
github.com/Defacto2/server/handler/app.Dirs.artifactEditor({{0xc00004e150, 0x23}, {0xc00004e00f, 0x23}, {0xc00004e191, 0x23}, {0xc0015556b7, 0x7}}, 0xc00242d408, 0xc000ddb500, ...)
	/home/ben/github/server/handler/app/dirs.go:182 +0x94d
github.com/Defacto2/server/handler/app.Dirs.Artifact({{0xc00004e150, 0x23}, {0xc00004e00f, 0x23}, {0xc00004e191, 0x23}, {0xc0015556b7, 0x7}}, {0x228ce78, 0xc001d18000}, ...)
	/home/ben/github/server/handler/app/dirs.go:105 +0x22a
github.com/Defacto2/server/handler.Configuration.website.func3({0x228ce78, 0xc001d18000})
	/home/ben/github/server/handler/router.go:208 +0xd1
github.com/labstack/echo/v4.(*Echo).add.func1({0x228ce78, 0xc001d18000})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:587 +0x4b
github.com/Defacto2/server/handler.Configuration.nonce.Middleware.MiddlewareWithConfig.func1.1({0x228ce78, 0xc001d18000})
	/home/ben/go/pkg/mod/github.com/labstack/echo-contrib@v0.17.1/session/session.go:73 +0x104
github.com/Defacto2/server/handler.Configuration.Controller.RemoveTrailingSlashWithConfig.func2.1({0x228ce78, 0xc001d18000})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/slash.go:117 +0x1fd
github.com/labstack/echo/v4/middleware.RequestLoggerConfig.ToMiddleware.func1.1({0x228ce78, 0xc001d18000})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/request_logger.go:286 +0x16b
github.com/Defacto2/server/handler.Configuration.Controller.Secure.SecureWithConfig.func4.1({0x228ce78, 0xc001d18000})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/secure.go:141 +0x364
github.com/labstack/echo/v4.(*Echo).ServeHTTP.func1({0x228ce78, 0xc001d18000})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:668 +0x127
github.com/Defacto2/server/handler.Configuration.Controller.NonWWWRedirect.NonWWWRedirectWithConfig.redirect.func6.1({0x228ce78, 0xc001d18000})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/redirect.go:152 +0xf3
github.com/labstack/echo/v4/middleware.RewriteWithConfig.func1.1({0x228ce78, 0xc001d18000})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/rewrite.go:77 +0x7f
github.com/labstack/echo/v4.(*Echo).ServeHTTP(0xc0001b2908, {0x2274ee8, 0xc00019f260}, 0xc0015ab200)
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:674 +0x327
net/http.serverHandler.ServeHTTP({0xc00068c210?}, {0x2274ee8?, 0xc00019f260?}, 0x6?)
	/usr/local/go/src/net/http/server.go:3137 +0x8e
net/http.(*conn).serve(0xc001610120, {0x2276fc8, 0xc001586360})
	/usr/local/go/src/net/http/server.go:2039 +0x5e8
created by net/http.(*Server).Serve in goroutine 11
	/usr/local/go/src/net/http/server.go:3285 +0x4b4

goroutine 1 [chan receive, 171 minutes]:
github.com/Defacto2/server/handler.(*Configuration).ShutdownHTTP(0xc0000460e3?, 0xc0001b2908, 0xc000088800)
	/home/ben/github/server/handler/handler.go:219 +0xa5
main.main()
	/home/ben/github/server/server.go:105 +0x66e

goroutine 6 [select]:
go.opencensus.io/stats/view.(*worker).start(0xc00022cf00)
	/home/ben/go/pkg/mod/go.opencensus.io@v0.24.0/stats/view/worker.go:292 +0x9f
created by go.opencensus.io/stats/view.init.0 in goroutine 1
	/home/ben/go/pkg/mod/go.opencensus.io@v0.24.0/stats/view/worker.go:34 +0x8d

goroutine 7 [select, 171 minutes]:
database/sql.(*DB).connectionOpener(0xc00042b790, {0x2277000, 0xc0001d3b30})
	/usr/local/go/src/database/sql/sql.go:1246 +0x87
created by database/sql.OpenDB in goroutine 1
	/usr/local/go/src/database/sql/sql.go:824 +0x14c

goroutine 11 [IO wait, 12 minutes]:
internal/poll.runtime_pollWait(0x7fec550c7e28, 0x72)
	/usr/local/go/src/runtime/netpoll.go:345 +0x85
internal/poll.(*pollDesc).wait(0x8?, 0xc0000979f8?, 0x0)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:84 +0x27
internal/poll.(*pollDesc).waitRead(...)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:89
internal/poll.(*FD).Accept(0xc001590000)
	/usr/local/go/src/internal/poll/fd_unix.go:611 +0x2ac
net.(*netFD).accept(0xc001590000)
	/usr/local/go/src/net/fd_unix.go:172 +0x29
net.(*TCPListener).accept(0xc000dd2060)
	/usr/local/go/src/net/tcpsock_posix.go:159 +0x1e
net.(*TCPListener).AcceptTCP(0xc000dd2060)
	/usr/local/go/src/net/tcpsock.go:314 +0x30
github.com/labstack/echo/v4.tcpKeepAliveListener.Accept({0x449ec0?})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:994 +0x17
net/http.(*Server).Serve(0xc0002321e0, {0x2275008, 0xc00158e018})
	/usr/local/go/src/net/http/server.go:3255 +0x33e
github.com/labstack/echo/v4.(*Echo).Start(0xc0001b2908, {0xc001580030, 0xe})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:691 +0xd2
github.com/Defacto2/server/handler.(*Configuration).StartHTTP(0xc00035e000, 0xc0001b2908, 0xc000088800)
	/home/ben/github/server/handler/handler.go:299 +0x6d
created by github.com/Defacto2/server/handler.(*Configuration).Start in goroutine 1
	/home/ben/github/server/handler/handler.go:280 +0x3ab

goroutine 14 [syscall, 171 minutes]:
os/signal.signal_recv()
	/usr/local/go/src/runtime/sigqueue.go:152 +0x29
os/signal.loop()
	/usr/local/go/src/os/signal/signal_unix.go:23 +0x13
created by os/signal.Notify.func1.1 in goroutine 1
	/usr/local/go/src/os/signal/signal.go:151 +0x1f

goroutine 18691 [IO wait]:
internal/poll.runtime_pollWait(0x7fec550c7380, 0x72)
	/usr/local/go/src/runtime/netpoll.go:345 +0x85
internal/poll.(*pollDesc).wait(0xc00341c200?, 0xc003427000?, 0x0)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:84 +0x27
internal/poll.(*pollDesc).waitRead(...)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:89
internal/poll.(*FD).Read(0xc00341c200, {0xc003427000, 0x1000, 0x1000})
	/usr/local/go/src/internal/poll/fd_unix.go:164 +0x27a
net.(*netFD).Read(0xc00341c200, {0xc003427000?, 0xc0015d1a98?, 0x4f0905?})
	/usr/local/go/src/net/fd_posix.go:55 +0x25
net.(*conn).Read(0xc001800010, {0xc003427000?, 0x0?, 0xc00160e0c8?})
	/usr/local/go/src/net/net.go:179 +0x45
net/http.(*connReader).Read(0xc00160e0c0, {0xc003427000, 0x1000, 0x1000})
	/usr/local/go/src/net/http/server.go:789 +0x14b
bufio.(*Reader).fill(0xc001761200)
	/usr/local/go/src/bufio/bufio.go:110 +0x103
bufio.(*Reader).Peek(0xc001761200, 0x4)
	/usr/local/go/src/bufio/bufio.go:148 +0x53
net/http.(*conn).serve(0xc003400240, {0x2276fc8, 0xc001586360})
	/usr/local/go/src/net/http/server.go:2074 +0x749
created by net/http.(*Server).Serve in goroutine 11
	/usr/local/go/src/net/http/server.go:3285 +0x4b4

goroutine 19622 [IO wait, 4 minutes]:
internal/poll.runtime_pollWait(0x7fec550c7668, 0x72)
	/usr/local/go/src/runtime/netpoll.go:345 +0x85
internal/poll.(*pollDesc).wait(0xc0027c8080?, 0xc0027c7000?, 0x0)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:84 +0x27
internal/poll.(*pollDesc).waitRead(...)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:89
internal/poll.(*FD).Read(0xc0027c8080, {0xc0027c7000, 0x1000, 0x1000})
	/usr/local/go/src/internal/poll/fd_unix.go:164 +0x27a
net.(*netFD).Read(0xc0027c8080, {0xc0027c7000?, 0xc003815a98?, 0x4f0905?})
	/usr/local/go/src/net/fd_posix.go:55 +0x25
net.(*conn).Read(0xc000450000, {0xc0027c7000?, 0x0?, 0xc0008a9b08?})
	/usr/local/go/src/net/net.go:179 +0x45
net/http.(*connReader).Read(0xc0008a9b00, {0xc0027c7000, 0x1000, 0x1000})
	/usr/local/go/src/net/http/server.go:789 +0x14b
bufio.(*Reader).fill(0xc000752c60)
	/usr/local/go/src/bufio/bufio.go:110 +0x103
bufio.(*Reader).Peek(0xc000752c60, 0x4)
	/usr/local/go/src/bufio/bufio.go:148 +0x53
net/http.(*conn).serve(0xc00019d440, {0x2276fc8, 0xc001586360})
	/usr/local/go/src/net/http/server.go:2074 +0x749
created by net/http.(*Server).Serve in goroutine 11
	/usr/local/go/src/net/http/server.go:3285 +0x4b4

goroutine 21818 [IO wait]:
internal/poll.runtime_pollWait(0x7fec550c7d30, 0x72)
	/usr/local/go/src/runtime/netpoll.go:345 +0x85
internal/poll.(*pollDesc).wait(0xc000882800?, 0xc001b4a281?, 0x0)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:84 +0x27
internal/poll.(*pollDesc).waitRead(...)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:89
internal/poll.(*FD).Read(0xc000882800, {0xc001b4a281, 0x1, 0x1})
	/usr/local/go/src/internal/poll/fd_unix.go:164 +0x27a
net.(*netFD).Read(0xc000882800, {0xc001b4a281?, 0xc0055b5748?, 0x46faf0?})
	/usr/local/go/src/net/fd_posix.go:55 +0x25
net.(*conn).Read(0xc000c7abe8, {0xc001b4a281?, 0xc0055b57b0?, 0xc000faf268?})
	/usr/local/go/src/net/net.go:179 +0x45
net/http.(*connReader).backgroundRead(0xc001b4a270)
	/usr/local/go/src/net/http/server.go:681 +0x37
created by net/http.(*connReader).startBackgroundRead in goroutine 19204
	/usr/local/go/src/net/http/server.go:677 +0xba

goroutine 19204 [runnable]:
github.com/h2non/filetype/matchers.NewMatcher(...)
	/home/ben/go/pkg/mod/github.com/h2non/filetype@v1.1.3/matchers/matchers.go:34
github.com/h2non/filetype.AddMatcher(...)
	/home/ben/go/pkg/mod/github.com/h2non/filetype@v1.1.3/match.go:68
github.com/Defacto2/server/handler/app.artifactMagic({0xc00121f090?, 0xc0008a9b60?})
	/home/ben/github/server/handler/app/dirs.go:580 +0x859
github.com/Defacto2/server/handler/app.Dirs.artifactEditor({{0xc00004e150, 0x23}, {0xc00004e00f, 0x23}, {0xc00004e191, 0x23}, {0xc001555777, 0x7}}, 0xc00242d808, 0xc0008a9b60, ...)
	/home/ben/github/server/handler/app/dirs.go:182 +0x94d
github.com/Defacto2/server/handler/app.Dirs.Artifact({{0xc00004e150, 0x23}, {0xc00004e00f, 0x23}, {0xc00004e191, 0x23}, {0xc001555777, 0x7}}, {0x228ce78, 0xc001c53860}, ...)
	/home/ben/github/server/handler/app/dirs.go:105 +0x22a
github.com/Defacto2/server/handler.Configuration.website.func3({0x228ce78, 0xc001c53860})
	/home/ben/github/server/handler/router.go:208 +0xd1
github.com/labstack/echo/v4.(*Echo).add.func1({0x228ce78, 0xc001c53860})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:587 +0x4b
github.com/Defacto2/server/handler.Configuration.nonce.Middleware.MiddlewareWithConfig.func1.1({0x228ce78, 0xc001c53860})
	/home/ben/go/pkg/mod/github.com/labstack/echo-contrib@v0.17.1/session/session.go:73 +0x104
github.com/Defacto2/server/handler.Configuration.Controller.RemoveTrailingSlashWithConfig.func2.1({0x228ce78, 0xc001c53860})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/slash.go:117 +0x1fd
github.com/labstack/echo/v4/middleware.RequestLoggerConfig.ToMiddleware.func1.1({0x228ce78, 0xc001c53860})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/request_logger.go:286 +0x16b
github.com/Defacto2/server/handler.Configuration.Controller.Secure.SecureWithConfig.func4.1({0x228ce78, 0xc001c53860})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/secure.go:141 +0x364
github.com/labstack/echo/v4.(*Echo).ServeHTTP.func1({0x228ce78, 0xc001c53860})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:668 +0x127
github.com/Defacto2/server/handler.Configuration.Controller.NonWWWRedirect.NonWWWRedirectWithConfig.redirect.func6.1({0x228ce78, 0xc001c53860})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/redirect.go:152 +0xf3
github.com/labstack/echo/v4/middleware.RewriteWithConfig.func1.1({0x228ce78, 0xc001c53860})
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/middleware/rewrite.go:77 +0x7f
github.com/labstack/echo/v4.(*Echo).ServeHTTP(0xc0001b2908, {0x2274ee8, 0xc00019f340}, 0xc0015abd40)
	/home/ben/go/pkg/mod/github.com/labstack/echo/v4@v4.12.0/echo.go:674 +0x327
net/http.serverHandler.ServeHTTP({0xc001b4a270?}, {0x2274ee8?, 0xc00019f340?}, 0x6?)
	/usr/local/go/src/net/http/server.go:3137 +0x8e
net/http.(*conn).serve(0xc003358090, {0x2276fc8, 0xc001586360})
	/usr/local/go/src/net/http/server.go:2039 +0x5e8
created by net/http.(*Server).Serve in goroutine 11
	/usr/local/go/src/net/http/server.go:3285 +0x4b4

goroutine 18595 [IO wait, 4 minutes]:
internal/poll.runtime_pollWait(0x7fec550c7b40, 0x72)
	/usr/local/go/src/runtime/netpoll.go:345 +0x85
internal/poll.(*pollDesc).wait(0xc001748000?, 0xc002a99000?, 0x0)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:84 +0x27
internal/poll.(*pollDesc).waitRead(...)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:89
internal/poll.(*FD).Read(0xc001748000, {0xc002a99000, 0x1000, 0x1000})
	/usr/local/go/src/internal/poll/fd_unix.go:164 +0x27a
net.(*netFD).Read(0xc001748000, {0xc002a99000?, 0xc0015d3a98?, 0x4f0905?})
	/usr/local/go/src/net/fd_posix.go:55 +0x25
net.(*conn).Read(0xc001758820, {0xc002a99000?, 0x0?, 0xc000caa038?})
	/usr/local/go/src/net/net.go:179 +0x45
net/http.(*connReader).Read(0xc000caa030, {0xc002a99000, 0x1000, 0x1000})
	/usr/local/go/src/net/http/server.go:789 +0x14b
bufio.(*Reader).fill(0xc00098c8a0)
	/usr/local/go/src/bufio/bufio.go:110 +0x103
bufio.(*Reader).Peek(0xc00098c8a0, 0x4)
	/usr/local/go/src/bufio/bufio.go:148 +0x53
net/http.(*conn).serve(0xc0059fc000, {0x2276fc8, 0xc001586360})
	/usr/local/go/src/net/http/server.go:2074 +0x749
created by net/http.(*Server).Serve in goroutine 11
	/usr/local/go/src/net/http/server.go:3285 +0x4b4

goroutine 18682 [IO wait, 4 minutes]:
internal/poll.runtime_pollWait(0x7fec550c7c38, 0x72)
	/usr/local/go/src/runtime/netpoll.go:345 +0x85
internal/poll.(*pollDesc).wait(0xc000f06800?, 0xc003470000?, 0x0)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:84 +0x27
internal/poll.(*pollDesc).waitRead(...)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:89
internal/poll.(*FD).Read(0xc000f06800, {0xc003470000, 0x1000, 0x1000})
	/usr/local/go/src/internal/poll/fd_unix.go:164 +0x27a
net.(*netFD).Read(0xc000f06800, {0xc003470000?, 0xc00381ba98?, 0x4f0905?})
	/usr/local/go/src/net/fd_posix.go:55 +0x25
net.(*conn).Read(0xc0017588f0, {0xc003470000?, 0x0?, 0xc0009ea788?})
	/usr/local/go/src/net/net.go:179 +0x45
net/http.(*connReader).Read(0xc0009ea780, {0xc003470000, 0x1000, 0x1000})
	/usr/local/go/src/net/http/server.go:789 +0x14b
bufio.(*Reader).fill(0xc0009dbda0)
	/usr/local/go/src/bufio/bufio.go:110 +0x103
bufio.(*Reader).Peek(0xc0009dbda0, 0x4)
	/usr/local/go/src/bufio/bufio.go:148 +0x53
net/http.(*conn).serve(0xc0059fc630, {0x2276fc8, 0xc001586360})
	/usr/local/go/src/net/http/server.go:2074 +0x749
created by net/http.(*Server).Serve in goroutine 11
	/usr/local/go/src/net/http/server.go:3285 +0x4b4

goroutine 18684 [IO wait]:
internal/poll.runtime_pollWait(0x7fec550c7858, 0x72)
	/usr/local/go/src/runtime/netpoll.go:345 +0x85
internal/poll.(*pollDesc).wait(0xc00160a080?, 0xc00162c000?, 0x0)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:84 +0x27
internal/poll.(*pollDesc).waitRead(...)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:89
internal/poll.(*FD).Read(0xc00160a080, {0xc00162c000, 0x1000, 0x1000})
	/usr/local/go/src/internal/poll/fd_unix.go:164 +0x27a
net.(*netFD).Read(0xc00160a080, {0xc00162c000?, 0xc00005da98?, 0x4f0905?})
	/usr/local/go/src/net/fd_posix.go:55 +0x25
net.(*conn).Read(0xc000632438, {0xc00162c000?, 0x0?, 0xc000faf268?})
	/usr/local/go/src/net/net.go:179 +0x45
net/http.(*connReader).Read(0xc000faf260, {0xc00162c000, 0x1000, 0x1000})
	/usr/local/go/src/net/http/server.go:789 +0x14b
bufio.(*Reader).fill(0xc001420180)
	/usr/local/go/src/bufio/bufio.go:110 +0x103
bufio.(*Reader).Peek(0xc001420180, 0x4)
	/usr/local/go/src/bufio/bufio.go:148 +0x53
net/http.(*conn).serve(0xc001610090, {0x2276fc8, 0xc001586360})
	/usr/local/go/src/net/http/server.go:2074 +0x749
created by net/http.(*Server).Serve in goroutine 11
	/usr/local/go/src/net/http/server.go:3285 +0x4b4

goroutine 18683 [IO wait]:
internal/poll.runtime_pollWait(0x7fec550c7a48, 0x72)
	/usr/local/go/src/runtime/netpoll.go:345 +0x85
internal/poll.(*pollDesc).wait(0xc00160a000?, 0xc001628000?, 0x0)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:84 +0x27
internal/poll.(*pollDesc).waitRead(...)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:89
internal/poll.(*FD).Read(0xc00160a000, {0xc001628000, 0x1000, 0x1000})
	/usr/local/go/src/internal/poll/fd_unix.go:164 +0x27a
net.(*netFD).Read(0xc00160a000, {0xc001628000?, 0xc001f45a98?, 0x4f0905?})
	/usr/local/go/src/net/fd_posix.go:55 +0x25
net.(*conn).Read(0xc000632430, {0xc001628000?, 0x0?, 0xc000c2a4e8?})
	/usr/local/go/src/net/net.go:179 +0x45
net/http.(*connReader).Read(0xc000c2a4e0, {0xc001628000, 0x1000, 0x1000})
	/usr/local/go/src/net/http/server.go:789 +0x14b
bufio.(*Reader).fill(0xc001420120)
	/usr/local/go/src/bufio/bufio.go:110 +0x103
bufio.(*Reader).Peek(0xc001420120, 0x4)
	/usr/local/go/src/bufio/bufio.go:148 +0x53
net/http.(*conn).serve(0xc001610000, {0x2276fc8, 0xc001586360})
	/usr/local/go/src/net/http/server.go:2074 +0x749
created by net/http.(*Server).Serve in goroutine 11
	/usr/local/go/src/net/http/server.go:3285 +0x4b4

goroutine 21816 [IO wait]:
internal/poll.runtime_pollWait(0x7fec550c7950, 0x72)
	/usr/local/go/src/runtime/netpoll.go:345 +0x85
internal/poll.(*pollDesc).wait(0xc00160a100?, 0xc00068c221?, 0x0)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:84 +0x27
internal/poll.(*pollDesc).waitRead(...)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:89
internal/poll.(*FD).Read(0xc00160a100, {0xc00068c221, 0x1, 0x1})
	/usr/local/go/src/internal/poll/fd_unix.go:164 +0x27a
net.(*netFD).Read(0xc00160a100, {0xc00068c221?, 0xc005ba8748?, 0x46faf0?})
	/usr/local/go/src/net/fd_posix.go:55 +0x25
net.(*conn).Read(0xc000632440, {0xc00068c221?, 0xc005ba87b0?, 0xc00068c218?})
	/usr/local/go/src/net/net.go:179 +0x45
net/http.(*connReader).backgroundRead(0xc00068c210)
	/usr/local/go/src/net/http/server.go:681 +0x37
created by net/http.(*connReader).startBackgroundRead in goroutine 18685
	/usr/local/go/src/net/http/server.go:677 +0xba
```

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