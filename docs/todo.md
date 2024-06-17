# TODOs and tasks

  * (star) __*__ prefix indicates a *low priority* task.
  * (question) __?__ prefix indicates an *idea* or *doubtful*.

### Terminal commands and flags

- [ ] *? Command to clean up the database and remove all orphaned records.
- [ ] *? Command to reindex the database, both to erase and rebuild the indexes.

### Files and assets

- [ ] Create a htmx, live _classifications page_ for editors, using the advanced uploader `<select>` fields.
- [ ] Mobile support for Data editor.
- [ ] Data editor button should reload the page when data Editor module is active.

### Menus and layout

- [ ] Create a menu link to DigitalOcean referal page, [or/and] add a link to the thanks page.
- [ ] Create a locked menu option to search the database by file ID or UUID or ~~URL~~.

### Database

- [ ] Create a DB fix to detect and rebadge msdos and windows trainers.
- [ ] `OrderBy` Name/Count /html3/groups? https://pkg.go.dev/sort#example-package-SortKeys

### Backend

- [ ] *? Implememnt a [scheduling library for Go](https://github.com/reugn/go-quartz)
- [ ] [xstrings](https://github.com/huandu/xstrings) for string manipulation.
- [ ] Errors cleanup, never return raw errors, always wrap them. And also never use, "xxx failed or broke" as an error message. Instead use "doing xxx" or "while doing xxx".