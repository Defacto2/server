package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bengarrett/df2023/router/html3"
	"github.com/labstack/echo/v4"
)

var (
	ErrEmpty    = errors.New("empty directory input")
	ErrNoReader = errors.New("reader cannot be nil, it should be os.stdin")
)

func ParsePsVersion(s string) string {
	if x := strings.Split(s, " "); len(x) > 2 {
		_, err := strconv.ParseFloat(x[1], 32)
		if err != nil {
			return fmt.Sprintln(s)
		}
		return fmt.Sprintf(" with %s\n", strings.Join(x[0:2], " "))
	}
	return fmt.Sprintln(s)
}

func CustomErrorHandler(err error, c echo.Context) {
	splitPaths := func(r rune) bool {
		return r == '/'
	}
	rel := strings.FieldsFunc(c.Path(), splitPaths)
	html3Route := len(rel) > 0 && rel[0] == "html3"
	if html3Route {
		if err := html3.Error(err, c); err != nil {
			panic(err) // TODO: logger?
		}
		return
	}
	code := http.StatusInternalServerError
	msg := "internal server error"
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = fmt.Sprint(he.Message)
	}
	c.Logger().Error(err)
	c.String(code, fmt.Sprintf("%d - %s", code, msg))
	// errorPage := fmt.Sprintf("%d.html", code)
	// if err := c.File(errorPage); err != nil {
	// 	c.Logger().Error(err)
	// }
}

// TODO: remove unused funcs below

// Prompt asks the user for a string configuration value and saves it.
func Prompt(keep string) string {
	fmt.Println(keep)
	s, err := String(os.Stdin)
	if errors.Is(err, ErrEmpty) {
		os.Exit(0)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return s
}

// String parses the reader, looking for a string and newline.
// Except for testing, r should always be os.Stdin.
func String(r io.Reader) (string, error) {
	if r == nil {
		return "", ErrNoReader
	}
	// allow multiple word user input
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		txt := scanner.Text()
		switch txt {
		case "":
			return "", ErrEmpty
		case "-":
			return "", nil
		default:
			return txt, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", ErrEmpty
}
