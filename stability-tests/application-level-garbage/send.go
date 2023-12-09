package main

import (
<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/app/appmessage"
	"github.com/zoomy-network/zoomyd/app/protocol/common"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/consensushashing"
	"github.com/zoomy-network/zoomyd/infrastructure/network/netadapter/standalone"
=======
>>>>>>> Stashed changes
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/app/appmessage"
	"github.com/zoomy-network/zoomyd/app/protocol/common"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/consensushashing"
	"github.com/zoomy-network/zoomyd/infrastructure/network/netadapter/standalone"
)

func sendBlocks(address string, minimalNetAdapter *standalone.MinimalNetAdapter, blocksChan <-chan *externalapi.DomainBlock) error {
	for block := range blocksChan {
		routes, err := minimalNetAdapter.Connect(address)
		if err != nil {
			return err
		}

		blockHash := consensushashing.BlockHash(block)
		log.Infof("Sending block %s", blockHash)

		err = routes.OutgoingRoute.Enqueue(&appmessage.MsgInvRelayBlock{
			Hash: blockHash,
		})
		if err != nil {
			return err
		}

		message, err := routes.WaitForMessageOfType(appmessage.CmdRequestRelayBlocks, common.DefaultTimeout)
		if err != nil {
			return err
		}
		requestRelayBlockMessage := message.(*appmessage.MsgRequestRelayBlocks)
		if len(requestRelayBlockMessage.Hashes) != 1 || *requestRelayBlockMessage.Hashes[0] != *blockHash {
			return errors.Errorf("Expecting requested hashes to be [%s], but got %v",
				blockHash, requestRelayBlockMessage.Hashes)
		}

		err = routes.OutgoingRoute.Enqueue(appmessage.DomainBlockToMsgBlock(block))
		if err != nil {
			return err
		}

		// TODO(libp2p): Wait for reject message once it has been implemented
		err = routes.WaitForDisconnect(common.DefaultTimeout)
		if err != nil {
			return err
		}
	}
	return nil
}
