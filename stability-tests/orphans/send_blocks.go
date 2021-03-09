package main

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/standalone"
	"github.com/pkg/errors"
)

func sendBlocks(routes *standalone.Routes, blocks []*externalapi.DomainBlock, topBlock *externalapi.DomainBlock) error {
	topBlockHash := consensushashing.BlockHash(topBlock)
	log.Infof("Sending top block with hash %s", topBlockHash)
	err := routes.OutgoingRoute.Enqueue(&appmessage.MsgInvRelayBlock{Hash: topBlockHash})
	if err != nil {
		return err
	}

	err = waitForRequestAndSend(routes, topBlock)
	if err != nil {
		return err
	}

	for i := len(blocks) - 1; i >= 0; i-- {
		block := blocks[i]

		orphanBlock := topBlock
		if i+1 != len(blocks) {
			orphanBlock = blocks[i+1]
		}
		log.Infof("Waiting for request for block locator for block number %d with hash %s", i, consensushashing.BlockHash(block))
		err = waitForRequestForBlockLocator(routes, orphanBlock)
		if err != nil {
			return err
		}

		log.Infof("Waiting for request and sending block number %d with hash %s", i, consensushashing.BlockHash(block))
		err = waitForRequestAndSend(routes, block)
		if err != nil {
			return err
		}
	}

	return nil
}

func waitForRequestForBlockLocator(routes *standalone.Routes, orphanBlock *externalapi.DomainBlock) error {
	message, err := routes.WaitForMessageOfType(appmessage.CmdRequestBlockLocator, timeout)
	if err != nil {
		return err
	}
	requestBlockLocatorMessage := message.(*appmessage.MsgRequestBlockLocator)

	orphanBlockHash := consensushashing.BlockHash(orphanBlock)
	if *requestBlockLocatorMessage.HighHash != *orphanBlockHash {
		return errors.Errorf("expected blockLocator request high hash to be %s but got %s",
			orphanBlockHash, requestBlockLocatorMessage.HighHash)
	}

	locator := appmessage.NewMsgBlockLocator([]*externalapi.DomainHash{orphanBlockHash, activeConfig().ActiveNetParams.GenesisHash})
	return routes.OutgoingRoute.Enqueue(locator)
}

func waitForRequestAndSend(routes *standalone.Routes, block *externalapi.DomainBlock) error {
	message, err := routes.WaitForMessageOfType(appmessage.CmdRequestRelayBlocks, timeout)
	if err != nil {
		return err
	}

	requestRelayBlockMessage := message.(*appmessage.MsgRequestRelayBlocks)

	blockHash := consensushashing.BlockHash(block)
	if len(requestRelayBlockMessage.Hashes) != 1 || *requestRelayBlockMessage.Hashes[0] != *blockHash {
		return errors.Errorf("expecting requested hashes to be [%s], but got %v",
			blockHash, requestRelayBlockMessage.Hashes)
	}

	return routes.OutgoingRoute.Enqueue(appmessage.DomainBlockToMsgBlock(block))
}
