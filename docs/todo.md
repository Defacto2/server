# TODOs and tasks

### Database

- [ ] Create PostgreSQL *indexes* with case-sensitive strings.
- [ ] Some form of database timeout.
- [ ] All SQL statements need a sign-in account to display records with `delete_at` ~ `qm.WithDeleted`
- [x] In the app, confirm PS_HOST_NAME: "host.docker.internal" and handle Docker differently with startup messages.


#### Automatic database corrections

- [x] `/g/damn-excellent-ansi-designers` > `damn-excellent-ansi-design`
- [x] `/g/the-original-funny-guys` > `originally-funny-guys`

### Pages

- [ ] Show a *custom error page when the file download* is missing or the root directory is broken.
- [ ] Run an *automated test to confirm 200 status* for all routes. Run this on startup using a defer func?
- [ ] Tests for routes and templates.

### Known broken links

- [x] http://localhost:1323/g/x_pression-design
- [x] http://localhost:1323/g/ice
- [x] http://localhost:1323/g/ansi-creators-in-demand
- [x] http://localhost:1323/g/nc_17
- [x] http://localhost:1323/g/share-and-enjoy
- [x] http://localhost:1323/g/north-american-pirate_phreak-association


### Possible TODOs

- [ ] `OrderBy` Name/Count /html3/groups?
https://pkg.go.dev/sort#example-package-SortKeys
- [ ] Move `OrderBy` params to cookies?
- [ ] (long) group/releaser pages should have a link to the end of the document.
- [ ] [model.Files.ListUpdates], rename the PSQL column from "updated_at" to "date_updated".
- [ ] Fetch the DOD nfo for w95, https://scenelist.org/nfo/DOD95C1H.ZIP

#### Support Unicode slug URLs as currently the regex removes all non alphanumeric chars.

```go
/*
Error:      	Not equal:
            	expected: "Mooñpeople"
            	actual  : "Moopeople"

				use utf8 lib to detect extended chars?
*/
```
