package main

import (
	"encoding/json"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/pkg/errors"
)

var blockBuffer []byte

func readBlocks() (<-chan *externalapi.DomainBlock, error) {
	c := make(chan *externalapi.DomainBlock)

	spawn("applicationLevelGarbage-readBlocks", func() {
		lineNum := 0
		for blockJSON := range common.ScanFile(activeConfig().BlocksFilePath) {
			domainBlock := &externalapi.DomainBlock{}

			err := json.Unmarshal(blockJSON, domainBlock)
			if err != nil {
				panic(errors.Wrapf(err, "error deserializing line No. %d with json %s", lineNum, blockJSON))
			}

			c <- domainBlock
		}
		close(c)
	})

	return c, nil
}
