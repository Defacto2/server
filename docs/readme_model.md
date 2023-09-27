# Model readme

The model directory is where the SQLBoiler QueryMods and building blocks exist. These blocks are interchangeable between different SQL database applications.

### [SQLBoiler readme](https://github.com/volatiletech/sqlboiler)

### Example tutorials

[Handling HTTP request in Go Echo framework](https://blog.boatswain.io/post/handling-http-request-in-go-echo-framework-1/)

[SQL in Go with SQLBoiler](https://thedevelopercafe.com/articles/sql-in-go-with-sqlboiler-ac8efc4c5cb8)

[Introduction to SQLBoiler: Go framework for ORMs](https://blog.logrocket.com/introduction-sqlboiler-go-framework-orms/)

---

### SQLBoiler example

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

### New clone table for testing data.

Create a new PostgreSQL, Files DB for unit testing.

`defacto2-tests` containing `files`, `groupnames`

```sh
sudo -u postgres -- createuser --createdb --pwprompt start
createdb --username start --password --owner start gettingstarted
psql --username start --password gettingstarted < schema.sql
```
