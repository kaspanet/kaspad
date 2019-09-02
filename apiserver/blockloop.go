package main

import (
	"github.com/daglabs/btcd/btcjson"
)

func blockLoop(client *apiServerClient, doneChan chan struct{}) error {
	selectedParentChain, blocks, err := collectCurrentSelectedParentChain(client)
	if err != nil {
		return err
	}
	log.Infof("aaaa %d %d", len(selectedParentChain), len(blocks))

loop:
	for {
		select {
		case blockAdded := <-client.onBlockAdded:
			log.Infof("blockAdded: %s", blockAdded.header)
		case chainChanged := <-client.onChainChanged:
			log.Infof("chainChanged: %+v", chainChanged)
		case <-doneChan:
			log.Infof("blockLoop stopped")
			break loop
		}
	}
	return nil
}

func collectCurrentSelectedParentChain(client *apiServerClient) ([]btcjson.ChainBlock, []btcjson.GetBlockVerboseResult, error) {
	var startHash *string
	var selectedParentChain []btcjson.ChainBlock
	var blocks []btcjson.GetBlockVerboseResult
	for {
		chainFromBlockResult, err := client.GetChainFromBlock(true, startHash)
		if err != nil {
			return nil, nil, err
		}

		if len(chainFromBlockResult.SelectedParentChain) == 0 {
			break
		}

		startHash = &chainFromBlockResult.SelectedParentChain[len(chainFromBlockResult.SelectedParentChain)-1].Hash
		selectedParentChain = append(selectedParentChain, chainFromBlockResult.SelectedParentChain...)
		blocks = append(blocks, chainFromBlockResult.Blocks...)
	}
	return selectedParentChain, blocks, nil
}
