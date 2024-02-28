# Golang changes

## v1.22
https://go.dev/blog/go1.22
https://tip.golang.org/doc/go1.22

New random package, [`math/rand/v2` package](https://pkg.go.dev/math/rand/v2).

```go
alpha := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
fmt.Println(alpha[rand.IntN(len(alpha))]) 

fmt.Println(rand.IntN(100))
```

New type `Null` for [database/sql package](https://pkg.go.dev/database/sql#Null).

```go
type Null[T any] struct {
	V     T
	Valid bool
}
```

New [`go/version` package](https://pkg.go.dev/go/version) for version comparison.

```go
func Compare(x, y string) int
func IsValid(x string) bool
func Lang(x string) string
```

## v1.21
https://go.dev/blog/go1.21
https://tip.golang.org/doc/go1.21


New built-in functions: min, max, clear.
https://go.dev/ref/spec#Min_and_max
https://go.dev/ref/spec#Clear


New slog package for structured logging.
https://pkg.go.dev/log/slog

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("hello", "count", 3)

slog.SetDefault(logger)

logger2 := logger.With("url", r.URL)

logger := slog.Default().With("id", systemID)
parserLogger := logger.WithGroup("parser")
parseInput(input, parserLogger)
```

New slices package for slice manipulation which are more efficient than `sort.`. // todo
https://pkg.go.dev/slices

```go
	smallInts := []int8{0, 42, -10, 8}
	slices.Sort(smallInts)
	fmt.Println(smallInts)
```

New maps package.
https://pkg.go.dev/maps

```go
maps.Clone
maps.Copy
maps.DeleteFunc
maps.Equal
maps.EqualFunc
```

New cmp package for comparing ordered values.
https://pkg.go.dev/cmp

```go
cmp.Compare
cmp.Less
cmp.Or
cmp.Order
```

## v1.20
https://go.dev/blog/go1.20
https://tip.golang.org/doc/go1.20

New function `errors.Join` for joining multiple errors into a single error.
https://pkg.go.dev/errors#Join

```go
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	err := errors.Join(err1, err2)
	fmt.Println(err)
	if errors.Is(err, err1) {
		fmt.Println("err is err1")
	}
	if errors.Is(err, err2) {
		fmt.Println("err is err2")
	}
```

