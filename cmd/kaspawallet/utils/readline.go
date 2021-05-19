package utils

import (
	"bufio"
	"strings"
)

// ReadLine reads one line from the given reader with trimmed white space.
func ReadLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(line)), nil
}
