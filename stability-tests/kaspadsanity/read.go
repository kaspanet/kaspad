package main

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func readArgs() <-chan []string {
	argsChan := make(chan []string)
	spawn("readArgs", func() {
		f, err := os.Open(cfg.CommandListFile)
		if err != nil {
			panic(errors.Wrapf(err, "error in Open"))
		}

		r := bufio.NewReader(f)
		for {
			line, _, err := r.ReadLine()

			if err == io.EOF {
				break
			}

			if err != nil {
				panic(errors.Wrapf(err, "error in ReadLine"))
			}

			trimmedLine := strings.TrimSpace(string(line))
			if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") {
				continue
			}

			argsChan <- strings.Split(trimmedLine, " ")
		}

		close(argsChan)
	})
	return argsChan
}
