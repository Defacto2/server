package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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
