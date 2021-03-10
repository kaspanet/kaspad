package common

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"os"

	"github.com/pkg/errors"
)

// ScanFile opens the file in the specified path, and returns a channel that
// sends the contents of the file line-by-line, ignoring lines beggining with //
func ScanFile(filePath string) <-chan []byte {
	c := make(chan []byte)

	spawn("ScanFile", func() {
		file, err := os.Open(filePath)
		if err != nil {
			panic(errors.Wrapf(err, "error opening file %s", filePath))
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				panic(errors.Wrap(err, "error reading line"))
			}

			line := scanner.Bytes()
			if bytes.HasPrefix(line, []byte("//")) {
				continue
			}

			c <- line
		}
		close(c)
	})

	return c
}

// ScanHexFile opens the file in the specified path, and returns a channel that
// sends the contents of the file line-by-line, ignoring lines beggining with //,
// parsing the hex data in all other lines
func ScanHexFile(filePath string) <-chan []byte {
	c := make(chan []byte)

	spawn("ScanHexFile", func() {
		lineNum := 1
		for lineHex := range ScanFile(filePath) {
			lineBytes := make([]byte, hex.DecodedLen(len(lineHex)))
			_, err := hex.Decode(lineBytes, lineHex)
			if err != nil {
				panic(errors.Wrapf(err, "error decoding line No. %d with hex %s", lineNum, lineHex))
			}

			c <- lineBytes

			lineNum++
		}
		close(c)
	})

	return c
}
