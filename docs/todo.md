# TODOs and tasks

  * (star) prefix indicates a *low priority* task.

### CLI

- [ ] * Command to cleanup the database?
- [ ] * Command to reindex the database? Both to erase and rebuild the indexes.

### Files

- [ ] Handle magazines with the file editor.
- [ ] Create a htmx live classifications page for editors, using the adv uploader <select> fields.

### Layout

- [ ] Create a menu link to DO referal page, or/and add a link to the thanks page.
- [ ] Create a locked menu option to search by database ID or UUID or ~~URL~~.

### Database

- [ ] Create DB fix to detect and rebadge msdos and windows trainers.
- [ ] Create PostgreSQL *indexes* with case-sensitive strings.
   https://wirekat.com/optimizing-sql-based-on-postgresql/
  | Partial Indexes: Use partial indexes when you only need to index a subset of rows, such as CREATE INDEX ON orders (order_date) WHERE status = 'SHIPPED';
  | Over-Indexing: Creating too many indexes can slow down write operations, as each index needs to be updated on INSERT, UPDATE, or DELETE.
  | Index Maintenance: Rebuild indexes periodically to deal with bloat using REINDEX.
  | Indexing Join Columns: Index columns that are used in JOIN conditions to improve join performance.
  > `combineGroup` and `(r Role) Distinct()`
- [ ] `OrderBy` Name/Count /html3/groups? https://pkg.go.dev/sort#example-package-SortKeys
- [ ] [model.Files.ListUpdates], rename the PSQL column from "updated_at" to "date_updated".

### Backend

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
