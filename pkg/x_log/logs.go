package x_log

import (
	"bufio"
	"fmt"
	"os"
)

// GetLogs reads and returns the last N lines from a file.
func GetLogs(filename string, n int) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return lines, err
	}

	// Return last N lines
	if len(lines) > n {
		return lines[len(lines)-n:], nil
	}
	return lines, nil
}

// PrintLogs prints each line with an optional colored prefix.
func PrintLogs(lines []string, prefix string, color func(a ...any) string) {
	for _, line := range lines {
		fmt.Println(color(prefix), line)
	}
}
