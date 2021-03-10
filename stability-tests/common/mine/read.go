package mine

import (
	"compress/gzip"
	"encoding/json"
	"os"
)

// JSONBlock is a json representation of a block in mine format
type JSONBlock struct {
	ID      string   `json:"id"`
	Parents []string `json:"parents"`
}

func readBlocks(jsonFile string) (<-chan JSONBlock, error) {
	f, err := os.Open(jsonFile)
	if err != nil {
		return nil, err
	}

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	defer gzipReader.Close()

	decoder := json.NewDecoder(gzipReader)

	blockChan := make(chan JSONBlock)
	spawn("mineFromJson.readBlocks", func() {
		// read open bracket
		_, err := decoder.Token()
		if err != nil {
			panic(err)
		}

		// while the array contains values
		for decoder.More() {
			var block JSONBlock
			// decode an array value (Message)
			err := decoder.Decode(&block)
			if err != nil {
				panic(err)
			}

			blockChan <- block
		}

		// read closing bracket
		_, err = decoder.Token()
		if err != nil {
			panic(err)
		}

		close(blockChan)
	})
	return blockChan, nil
}
