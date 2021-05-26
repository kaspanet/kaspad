package utils

import (
	"bufio"
	"strings"

	"github.com/pkg/errors"
)

// ReadLine reads one line from the given reader with trimmed white space.
func ReadLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return "", errors.WithStack(err)
	}

	return strings.TrimSpace(string(line)), nil
}
