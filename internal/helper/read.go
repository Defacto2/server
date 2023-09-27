package helper

import (
	"bufio"
	"fmt"
	"os"
)

func ReadFile(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}

	// read the file line by line using a scanner
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	text := []string{}
	for scanner.Scan() {
		x := scanner.Text()
		text = append(text, x)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	file.Close()

	for i, line := range text {
		fmt.Println(i, string(line))
	}

	return nil
}
