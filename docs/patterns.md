# Patterns

## New

This is a collection of patterns and idioms that could be useful with the new releases of Go.

- https://go.dev/blog/go1.22
- https://tip.golang.org/doc/go1.22
- https://go.dev/blog/go1.21
- https://tip.golang.org/doc/go1.21

### Randomization in Go 1.22

New random package, [`math/rand/v2` package](https://pkg.go.dev/math/rand/v2).

```go
// print a random letter
alpha := []string{
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", 
	"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
fmt.Println(alpha[rand.IntN(len(alpha))]) 

// print a random number between 0 and 100
fmt.Println(rand.IntN(100))
```

### New Null type in Go 1.22

New type `Null` for [database/sql package](https://pkg.go.dev/database/sql#Null).

```go
var s Null[string]
err := db.QueryRow("SELECT name FROM foo WHERE id=?", id).Scan(&s)
...
if s.Valid {
   // use s.V
} else {
   // NULL value
}
```

### New version comparison in Go 1.22

New [`go/version` package](https://pkg.go.dev/go/version) for version comparison.

```go
func Compare(x, y string) int
func IsValid(x string) bool
func Lang(x string) string
```

### New built-in functions: min, max, clear in Go 1.21

```go
x, y := 1, 100
m := min(x, y)
M := max(x, y)
c := max(5.0, y, 10)

alpha := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m"}
clear(alpha)
```

### New slices package in Go 1.21

New slices package for slice manipulation which are more efficient than `sort.`.
https://pkg.go.dev/slices

```go
smallInts := []int8{0, 42, -10, 8}
slices.Sort(smallInts)
fmt.Println(smallInts)
// Output: [-10 0 8 42]
```

## Database queries

The web application relies on an Object-relational mapping (ORM) implementation provided by [SQLBoiler](https://github.com/volatiletech/sqlboiler) to simplify development.

Some tutorials for SQLBoiler include:

- [SQL in Go with SQLBoiler](https://thedevelopercafe.com/articles/sql-in-go-with-sqlboiler-ac8efc4c5cb8)
- [Introduction to SQLBoiler: Go framework for ORMs](https://blog.logrocket.com/introduction-sqlboiler-go-framework-orms/)

---

Note, because I am lazy, the most of the descriptions below were generated using CoPilot AI.

```go
import(
	"github.com/Defacto2/server/internal/postgres/models"
)

got, err := models.Files().One(ctx, db)
if err != nil {
	panic(err)
}
fmt.Println("got record:", got.ID)
```

The above code snippet demonstrates how to query a single record from the `files` table. The `ctx` variable is a context.Context object and `db` is a *sql.DB object.

#### Find a record by ID

```go
f := &models.File{ID: 1}
found, err := models.FindFile(ctx, db, f.ID)
if err != nil {
	log.DPanic(err)
}
fmt.Println("found record:", found.Filename)
```

The above code snippet demonstrates how to query a single record from the `files` table using a specific ID.

#### Find records by ID

```go
import(
	"github.com/Defacto2/server/internal/postgres/models"
    "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

found, err := models.Files(qm.Where("id = ?", 1)).One(ctx, db)
if err != nil {
	log.DPanic(err)
}
fmt.Println("found record:", found.Filename)
```

The above code snippet also demonstrates how to query a single record from the `files` table using a specific ID. In this case, the query uses a query mod where clause.

#### Check if a record exists

```go
exists, err := models.FileExists(ctx, db, 1)
if err != nil {
	log.DPanic(err)
}
fmt.Println("found again exists?:", exists)
```

The above code snippet demonstrates how to check if a record exists in the `files` table.

#### Count records <small>with deleted</small>

```go
count, err := models.Files(qm.WithDeleted()).Count(ctx, db)
if err != nil {
	log.DPanic(err)
}
countPub, err := models.Files().Count(ctx, db)
if err != nil {
	log.DPanic(err)
}
fmt.Println("total files vs total files that are public:", count, "vs", countPub)
```

The above code snippet demonstrates how to count the number of records in the `files` table. The first query counts all records, including those that have been soft-deleted. The second query counts only records that are not soft-deleted.

#### Raw SQL queries

```go
var users []*models.File
err := models.Files(qm.SQL(`select * from files;`)).Bind(ctx, db, &users)
if err != nil {
	log.Fatal(err)
}
// range and print the results
fmt.Print("raw files:")
for i, user := range users {
	if i != 0 {
		fmt.Print(", ")
	}
	fmt.Print(user.Filename.String)
}
```

The above code snippet demonstrates how to execute a raw SQL query using SQLBoiler and bind the results to a slice of models.

```go
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
```

The above code snippet demonstrates how to execute a raw SQL query and bind the results to a slice of models. The `queries.Bind` function is provided by SQLBoiler.
