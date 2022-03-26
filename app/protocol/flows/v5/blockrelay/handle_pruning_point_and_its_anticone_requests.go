package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"sync/atomic"
)

// PruningPointAndItsAnticoneRequestsContext is the interface for the context needed for the HandlePruningPointAndItsAnticoneRequests flow.
type PruningPointAndItsAnticoneRequestsContext interface {
	Domain() domain.Domain
	Config() *config.Config
}

var isBusy uint32

// HandlePruningPointAndItsAnticoneRequests listens to appmessage.MsgRequestPruningPointAndItsAnticone messages and sends
// the pruning point and its anticone to the requesting peer.
func HandlePruningPointAndItsAnticoneRequests(context PruningPointAndItsAnticoneRequestsContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peerpkg.Peer) error {

	for {
		err := func() error {
			_, err := incomingRoute.Dequeue()
			if err != nil {
				return err
			}

			if !atomic.CompareAndSwapUint32(&isBusy, 0, 1) {
				return protocolerrors.Errorf(false, "node is busy with other pruning point anticone requests")
			}
			defer atomic.StoreUint32(&isBusy, 0)

			log.Debugf("Got request for pruning point and its anticone from %s", peer)

			pruningPointHeaders, err := context.Domain().Consensus().PruningPointHeaders()
			if err != nil {
				return err
			}

			log.Criticalf("Pruning point anticone size is %d", len(pruningPointHeaders))

			msgPruningPointHeaders := make([]*appmessage.MsgBlockHeader, len(pruningPointHeaders))
			for i, header := range pruningPointHeaders {
				msgPruningPointHeaders[i] = appmessage.DomainBlockHeaderToBlockHeader(header)
			}

			err = outgoingRoute.Enqueue(appmessage.NewMsgPruningPoints(msgPruningPointHeaders))
			if err != nil {
				return err
			}

			pointAndItsAnticone, err := context.Domain().Consensus().PruningPointAndItsAnticone()
			if err != nil {
				return err
			}

			windowSize := context.Config().NetParams().DifficultyAdjustmentWindowSize
			daaWindowBlocks := make([]*externalapi.TrustedDataDataDAAHeader, 0, windowSize)
			daaWindowHashesToIndex := make(map[externalapi.DomainHash]int, windowSize)
			trustedDataDAABlockIndexes := make(map[externalapi.DomainHash][]uint64)

			ghostdagData := make([]*externalapi.BlockGHOSTDAGDataHashPair, 0)
			ghostdagDataHashToIndex := make(map[externalapi.DomainHash]int)
			trustedDataGHOSTDAGDataIndexes := make(map[externalapi.DomainHash][]uint64)
			for _, blockHash := range pointAndItsAnticone {
				blockDAAWindowHashes, err := context.Domain().Consensus().BlockDAAWindowHashes(blockHash)
				if err != nil {
					return err
				}

				trustedDataDAABlockIndexes[*blockHash] = make([]uint64, 0, windowSize)
				for i, daaBlockHash := range blockDAAWindowHashes {
					index, exists := daaWindowHashesToIndex[*daaBlockHash]
					if !exists {
						trustedDataDataDAAHeader, err := context.Domain().Consensus().TrustedDataDataDAAHeader(blockHash, daaBlockHash, uint64(i))
						if err != nil {
							return err
						}
						daaWindowBlocks = append(daaWindowBlocks, trustedDataDataDAAHeader)
						index = len(daaWindowBlocks) - 1
						daaWindowHashesToIndex[*daaBlockHash] = index
					}

					trustedDataDAABlockIndexes[*blockHash] = append(trustedDataDAABlockIndexes[*blockHash], uint64(index))
				}

				ghostdagDataBlockHashes, err := context.Domain().Consensus().TrustedBlockAssociatedGHOSTDAGDataBlockHashes(blockHash)
				if err != nil {
					return err
				}

				trustedDataGHOSTDAGDataIndexes[*blockHash] = make([]uint64, 0, context.Config().NetParams().K)
				for _, ghostdagDataBlockHash := range ghostdagDataBlockHashes {
					index, exists := ghostdagDataHashToIndex[*ghostdagDataBlockHash]
					if !exists {
						data, err := context.Domain().Consensus().TrustedGHOSTDAGData(ghostdagDataBlockHash)
						if err != nil {
							return err
						}
						ghostdagData = append(ghostdagData, &externalapi.BlockGHOSTDAGDataHashPair{
							Hash:         ghostdagDataBlockHash,
							GHOSTDAGData: data,
						})
						index = len(ghostdagData) - 1
						ghostdagDataHashToIndex[*ghostdagDataBlockHash] = index
					}

					trustedDataGHOSTDAGDataIndexes[*blockHash] = append(trustedDataGHOSTDAGDataIndexes[*blockHash], uint64(index))
				}
			}

			err = outgoingRoute.Enqueue(appmessage.DomainTrustedDataToTrustedData(daaWindowBlocks, ghostdagData))
			if err != nil {
				return err
			}

			for i, blockHash := range pointAndItsAnticone {
				block, err := context.Domain().Consensus().GetBlock(blockHash)
				if err != nil {
					return err
				}

				err = outgoingRoute.Enqueue(appmessage.DomainBlockWithTrustedDataToBlockWithTrustedDataV4(block, trustedDataDAABlockIndexes[*blockHash], trustedDataGHOSTDAGDataIndexes[*blockHash]))
				if err != nil {
					return err
				}

				if (i+1)%ibdBatchSize == 0 {
					// No timeout here, as we don't care if the syncee takes its time computing,
					// since it only blocks this dedicated flow
					message, err := incomingRoute.Dequeue()
					if err != nil {
						return err
					}
					if _, ok := message.(*appmessage.MsgRequestNextPruningPointAndItsAnticoneBlocks); !ok {
						return protocolerrors.Errorf(true, "received unexpected message type. "+
							"expected: %s, got: %s", appmessage.CmdRequestNextPruningPointAndItsAnticoneBlocks, message.Command())
					}
				}
			}

			err = outgoingRoute.Enqueue(appmessage.NewMsgDoneBlocksWithTrustedData())
			if err != nil {
				return err
			}

			log.Debugf("Sent pruning point and its anticone to %s", peer)
			return nil
		}()
		if err != nil {
			return err
		}
	}
}
