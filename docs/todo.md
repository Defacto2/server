# TODOs and tasks

  * (star) prefix indicates a low priority task.

### Files

- [ ] Handle magazines with the file editor.

### Layout

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

### URL /v/xxx pattern tests.

- [X] detect if the file contains /r/n or /n and replace with /n only.
		example, http://localhost:1323/v/af18f9b
- [X] detect if the file uses cp437 or unicode and convert to utf8.
        example, http://localhost:1323/v/b01de5b 
		         http://localhost:1323/v/b521c83
				 http://localhost:1323/v/b8297cf
