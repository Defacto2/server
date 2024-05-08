# TODOs and tasks

  * (star) prefix indicates a low priority task.

### URL /v/xxx pattern fixes.

- [ ] detect if the file contains /r/n or /n and replace with /n only.
		example, http://localhost:1323/v/af18f9b
- [ ] detect if the file uses cp437 or unicode and convert to utf8.
        example, http://localhost:1323/v/b01de5b 
		         http://localhost:1323/v/b521c83
				 http://localhost:1323/v/b8297cf

### Files

- [ ] Listing files for approval should have a colored border.
		example, http://localhost:1323/editor/for-approval
- [ ] Listing files for approval should have a colored border.
		example, http://localhost:1323/editor/deletions
- [ ] Listing unwanted files should have a colored border.
		example, http://localhost:1323/editor/unwanted
- [ ] Viewing artifact pages should have colored background?
- [ ] Handle magazines with the file editor.
- [ ] GitHub repo should always trim single forward slash, both prefix and suffix, for save and fix.
- [ ] 16colors should always trim single forward slash, both prefix and suffix, for save and fix.
- [ ] List relationships should reverse all the IDs and confirm they're valid int Ids.

### Layout

- [X] Uploader menu alignment is cut off when on a resized, half-width browser window.
- [X] Show a *custom error page when the file download* is missing or the root directory is broken.
- [ ] Textfile content using Open Sans should wrap any preformatted text to the width of the page, example, http://localhost:1323/f/a1377e
- [ ] * (long) group/releaser pages should have a link to the end of the document.

### Database

- [ ] Create DB fix to detect and rebadge msdos and windows trainers.
- [ ] Create PostgreSQL *indexes* with case-sensitive strings.
- [ ] Some form of database timeout.
- [ ] All SQL statements need a sign-in account to display records with `delete_at` ~ `qm.WithDeleted`
- [ ] `OrderBy` Name/Count /html3/groups? https://pkg.go.dev/sort#example-package-SortKeys
- [ ] [model.Files.ListUpdates], rename the PSQL column from "updated_at" to "date_updated".

### Backend

- [ ] Tests for routes and templates.
- [ ] * [Implememnt a sheduling library for Go](https://github.com/reugn/go-quartz)

#### Support Unicode slug URLs as currently the regex removes all non alphanumeric chars.

```go
/*
Error:      	Not equal:
            	expected: "Mooñpeople"
            	actual  : "Moopeople"

				use utf8 lib to detect extended chars?
*/
```
