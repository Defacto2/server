package helper

import (
	"bufio"
	"os"
)

func Lines(name string) (int, error) {
	file, err := os.Open(name)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return lines, nil
}
