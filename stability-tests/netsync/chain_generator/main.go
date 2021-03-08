package main

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/kaspanet/automation/stability-tests/common"
	"github.com/kaspanet/automation/stability-tests/common/mine"
	"github.com/pkg/errors"
)

func main() {
	err := parseConfig()
	if err != nil {
		panic(errors.Wrap(err, "error in parseConfig"))
	}
	common.UseLogger(backendLog, log.Level())

	blocks := generateBlocks()
	err = writeJSONToFile(blocks, cfg.TargetFile)
	if err != nil {
		panic(errors.Wrap(err, "error in generateBlocks"))
	}
}

func generateBlocks() []mine.JSONBlock {
	numBlocks := int(activeConfig().NumberOfBlocks)
	blocks := make([]mine.JSONBlock, 0, numBlocks)
	blocks = append(blocks, mine.JSONBlock{
		ID: "0",
	})
	for i := 1; i < numBlocks; i++ {
		blocks = append(blocks, mine.JSONBlock{
			ID:      strconv.Itoa(i),
			Parents: []string{strconv.Itoa(i - 1)},
		})
	}

	return blocks
}

func writeJSONToFile(blocks []mine.JSONBlock, fileName string) error {
	f, err := openFile(fileName)
	if err != nil {
		return errors.Wrap(err, "error in openFile")
	}
	encoder := json.NewEncoder(f)
	err = encoder.Encode(blocks)
	return errors.Wrap(err, "error in Encode")
}

func openFile(name string) (*os.File, error) {
	os.Remove(name)
	f, err := os.Create(name)
	return f, errors.WithStack(err)
}
