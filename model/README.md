# Model README

The model directory is where SQL Boiler Query Mods and Building blocks exist, these are interchangeable between different databases.

---

### [SQL Boiler README](https://github.com/volatiletech/sqlboiler)

[Example API in Echo and SQL Boiler](https://blog.boatswain.io/post/handling-http-request-in-go-echo-framework-1/)

---

### SQL Boiler examples

```go
	got, err := models.Files().One(ctx, db)
	if err != nil {
		log.DPanic(err)
	}
	fmt.Println("got record:", got.ID)
	f := &models.File{ID: 1}
	found, err := models.FindFile(ctx, db, f.ID)
	if err != nil {
		log.DPanic(err)
	}
	fmt.Println("found record:", found.Filename)
	foundAgain, err := models.Files(qm.Where("id = ?", 1)).One(ctx, db)
	if err != nil {
		log.DPanic(err)
	}
	fmt.Println("found again record:", foundAgain.Filename)
	exists, err := models.FileExists(ctx, db, foundAgain.ID)
	if err != nil {
		log.DPanic(err)
	}
	fmt.Println("found again exists?:", exists)
	count, err := models.Files(qm.WithDeleted()).Count(ctx, db)
	if err != nil {
		log.DPanic(err)
	}
	countPub, err := models.Files().Count(ctx, db)
	if err != nil {
		log.DPanic(err)
	}
	fmt.Println("total files vs total files that are public:", count, "vs", countPub)

	var users []*models.File
	rows, err := db.QueryContext(context.Background(), `select * from files;`)
	if err != nil {
		log.Fatal(err)
	}
	err = queries.Bind(rows, &users)
	if err != nil {
		log.Fatal(err)
	}
	rows.Close()

	users = nil
	err = models.Files(qm.SQL(`select * from files;`)).Bind(ctx, db, &users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("raw files:")
	for i, u := range users {
		if i != 0 {
			fmt.Print(", ")
		}
		fmt.Print(u.Filename.String)
	}
	fmt.Println()
```

---

### Database migrations from MySQL to Postgres.

- Rename `Files` table to `Release` or `Releases`
- Create a `Release_tests` table with a selection of 20 read-only records.
- Rename `files.createdat` etc to `?_at` aka `create_at`.

[PostgreSQL datatypes](https://www.postgresql.org/docs/current/datatype.html)

`CITEXT` type for case-insensitive character strings.

`files.filesize` should be converted to an `integer`, 4 bytes to permit a 2.147GB value.

`files.id` should be converted to a `serial` type.

There is no performance improvement for fixed-length, padded character types, meaning strings can use `varchar`(n) or `text`.

---

### UUID

`files.UUID` have be renamed from CFML style to the universal RFC-4122 syntax.

This will require the modification of queries when dealing with `/files/[uuid|000|4000]`.

CFML is 35 characters, 8-4-4-16.
`xxxxxxxx-xxxx-xxxx-xxxxxxxxxxxxxxxx`

RFC is 36 characters, 8-4-4-4-12.
`xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`

---

### Store NFOs and texts

We can store NFO and textfiles plus file_id.diz in database using the `bytea` hex-format, binary data type. It is more performant than the binary escape format.

https://www.postgresql.org/docs/current/datatype-binary.html

---

### Full text seach types.

https://www.postgresql.org/docs/current/datatype-textsearch.html

---