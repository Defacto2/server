# TODOs and tasks

### Database

- [ ] Create PostgreSQL *indexes* with case-sensitive strings.
- [ ] Some form of database timeout.
- [ ] All SQL statements need a sign-in account to display records with `delete_at` ~ `qm.WithDeleted`
- [x] In the app, confirm PS_HOST_NAME: "host.docker.internal" and handle Docker differently with startup messages.

### Pages

- [ ] Show a *custom error page when the file download* is missing or the root directory is broken.
- [ ] Tests for routes and templates.


### Possible TODOs

- [ ] `OrderBy` Name/Count /html3/groups? https://pkg.go.dev/sort#example-package-SortKeys
- [ ] (long) group/releaser pages should have a link to the end of the document.
- [ ] [model.Files.ListUpdates], rename the PSQL column from "updated_at" to "date_updated".
- [x] Fetch the DOD nfo for w95, https://scenelist.org/nfo/DOD95C1H.ZIP
- [ ] ~~Move `OrderBy` params to cookies?~~
- [ ] ~~Run an *automated test to confirm 200 status* for all routes. Run this on startup using a defer func?~~
- [ ] [Implememnt a sheduling library for Go](https://github.com/reugn/go-quartz)

#### Support Unicode slug URLs as currently the regex removes all non alphanumeric chars.

```go
/*
Error:      	Not equal:
            	expected: "Mooñpeople"
            	actual  : "Moopeople"

				use utf8 lib to detect extended chars?
*/
```
