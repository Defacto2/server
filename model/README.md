# Model README

The model directory is where SQL Boiler Query Mods and Building blocks exist, these are interchangeable between different databases.

---

### [SQL Boiler README](https://github.com/volatiletech/sqlboiler)

[PGLoader docs](https://pgloader.readthedocs.io/)

[Example API in Echo and SQL Boiler](https://blog.boatswain.io/post/handling-http-request-in-go-echo-framework-1/)

[SQL in Go with SQLBoiler](https://thedevelopercafe.com/articles/sql-in-go-with-sqlboiler-ac8efc4c5cb8)

[Introduction to SQLBoiler: Go framework for ORMs](https://blog.logrocket.com/introduction-sqlboiler-go-framework-orms/)

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

### Extending SQL Boiler's generated models

https://github.com/volatiletech/sqlboiler#extending-generated-models



```go
# method 1: simple funcs

// Package modext is for SQLBoiler helper methods
package modext

// UserFirstTimeSetup is an extension of the user model.
func UserFirstTimeSetup(ctx context.Context, db *sql.DB, u *models.User) error { ... }

# calling code

user, err := Users().One(ctx, db)
// elided error check

err = modext.UserFirstTimeSetup(ctx, db, user)
// elided error check
```

---

### New clone table for testing data.

Create a new PostgreSQL, Files DB for unit testing.

`defacto2-tests` containing `files`, `groupnames`

```sh
sudo -u postgres -- createuser --createdb --pwprompt start
createdb --username start --password --owner start gettingstarted
psql --username start --password gettingstarted < schema.sql
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

### Files content relationship table

Create a relationship files table that contains the filename content within of every archive release. 

We could also include columns containing size in bytes, sha256 hash, text body for full text searching. 

This would replace the `file_zip_content` column and also, create a sep CLI tool to scan the archives to fill out this data. For saftey and code maintenance, the tool needs to be sep from `server` and `df2` applications.