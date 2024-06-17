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
- [ ] Fix PNG binary being displayed in the text viewer. http://localhost:1323/f/af20fcb
- [ ] Fix broken Unicode multibyte text being displayed in the text viewer. http://localhost:1323/f/b12d05f

### Menus and layout

- [ ] Create a menu link to DigitalOcean referal page, [or/and] add a link to the thanks page.
- [ ] Create a locked menu option to search the database by file ID or UUID or ~~URL~~.
- [ ] Create a locked page with links for various file items that use unique website features,
      for example, DOS emulation for different archives file types and raw executables.
      And also different text file types for the text viewer.

### Database

- [ ] Create a DB fix to detect and rebadge msdos and windows trainers.
- [ ] `OrderBy` Name/Count /html3/groups? https://pkg.go.dev/sort#example-package-SortKeys

### Backend

- [ ] *? Implememnt a [scheduling library for Go](https://github.com/reugn/go-quartz)
- [ ] [xstrings](https://github.com/huandu/xstrings) for string manipulation.
- [ ] Errors cleanup, never return raw errors, always wrap them. And also never use, "xxx failed or broke" as an error message. Instead use "doing xxx" or "while doing xxx".


#### Future locked file items list for testing features.

- Unknown codepage: http://localhost:1323/f/ac2319e,http://localhost:1323/f/b0269ca [comparison:http://localhost:1323/f/ac1d9d3],
- GIF image: http://localhost:1323/f/b828636,http://localhost:1323/f/b42e22b,http://localhost:1323/f/ae2a407,
- Excess tail whitespace: http://localhost:1323/f/b830654,
- Missing newlines, requires wrap: http://localhost:1323/f/b14bb1,http://localhost:1323/f/b12fe37,http://localhost:1323/f/ad23d9c,http://localhost:1323/f/b122787,
- ?No text preview?: http://localhost:1323/f/af31a9,
- HTML file preview: http://localhost:1323/f/a722b1f,
- PDF file preview: http://localhost:1323/f/b04139,
- Block text file: http://localhost:1323/f/ad217af,http://localhost:1323/f/ae2a9cc,http://localhost:1323/f/ad2b193,
- [REQUIRES FIX] Multibyte Unicode: http://localhost:1323/f/b12d05f,http://localhost:1323/f/b53028e,
- IRL Link to: http://localhost:1323/f/b029330,http://localhost:1323/f/ba4805,http://localhost:1323/f/ab27f81,http://localhost:1323/f/b029330,
- href in text viewer: http://localhost:1323/f/a92c1dc,http://localhost:1323/f/a734e9,http://localhost:1323/f/ac2a79,
- [REQUIRES FIX] JSDOS unsupported zip archive: http://localhost:1323/f/a22af8,
- CP437 text pattern detection: http://localhost:1323/f/ab2f2b4,
- Unicode single byte: http://localhost:1323/f/a5191c3,http://localhost:1323/v/ab1fc8b,http://localhost:1323/f/b61f24f,
- Text viewer attempting to preview PNG image due to category: http://localhost:1323/f/af20fcb,
- Maximum download permitted, 1GB: http://localhost:1323/f/aa256f1,