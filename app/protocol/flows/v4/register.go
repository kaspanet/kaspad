package v4

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/flowcontext"
	"github.com/kaspanet/kaspad/app/protocol/flows/v4/addressexchange"
	"github.com/kaspanet/kaspad/app/protocol/flows/v4/blockrelay"
	"github.com/kaspanet/kaspad/app/protocol/flows/v4/ping"
	"github.com/kaspanet/kaspad/app/protocol/flows/v4/rejects"
	"github.com/kaspanet/kaspad/app/protocol/flows/v4/transactionrelay"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

type protocolManager interface {
	RegisterFlow(name string, router *routerpkg.Router, messageTypes []appmessage.MessageCommand, isStopping *uint32,
		errChan chan error, initializeFunc common.FlowInitializeFunc) *common.Flow
	RegisterOneTimeFlow(name string, router *routerpkg.Router, messageTypes []appmessage.MessageCommand,
		isStopping *uint32, stopChan chan error, initializeFunc common.FlowInitializeFunc) *common.Flow
	RegisterFlowWithCapacity(name string, capacity int, router *routerpkg.Router,
		messageTypes []appmessage.MessageCommand, isStopping *uint32,
		errChan chan error, initializeFunc common.FlowInitializeFunc) *common.Flow
	Context() *flowcontext.FlowContext
}

// Register is used in order to register all the protocol flows to the given router.
func Register(m protocolManager, router *routerpkg.Router, errChan chan error, isStopping *uint32) (flows []*common.Flow) {
	flows = registerAddressFlows(m, router, isStopping, errChan)
	flows = append(flows, registerBlockRelayFlows(m, router, isStopping, errChan)...)
	flows = append(flows, registerPingFlows(m, router, isStopping, errChan)...)
	flows = append(flows, registerTransactionRelayFlow(m, router, isStopping, errChan)...)
	flows = append(flows, registerRejectsFlow(m, router, isStopping, errChan)...)

	return flows
}

func registerAddressFlows(m protocolManager, router *routerpkg.Router, isStopping *uint32, errChan chan error) []*common.Flow {
	outgoingRoute := router.OutgoingRoute()

	return []*common.Flow{
		m.RegisterFlow("SendAddresses", router, []appmessage.MessageCommand{appmessage.CmdRequestAddresses}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return addressexchange.SendAddresses(m.Context(), incomingRoute, outgoingRoute)
			},
		),

		m.RegisterOneTimeFlow("ReceiveAddresses", router, []appmessage.MessageCommand{appmessage.CmdAddresses}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return addressexchange.ReceiveAddresses(m.Context(), incomingRoute, outgoingRoute, peer)
			},
		),
	}
}

func registerBlockRelayFlows(m protocolManager, router *routerpkg.Router, isStopping *uint32, errChan chan error) []*common.Flow {
	outgoingRoute := router.OutgoingRoute()

	return []*common.Flow{
		m.RegisterOneTimeFlow("SendVirtualSelectedParentInv", router, []appmessage.MessageCommand{},
			isStopping, errChan, func(route *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.SendVirtualSelectedParentInv(m.Context(), outgoingRoute, peer)
			}),

		m.RegisterFlow("HandleRelayInvs", router, []appmessage.MessageCommand{
			appmessage.CmdInvRelayBlock, appmessage.CmdBlock, appmessage.CmdBlockLocator,
			appmessage.CmdDoneHeaders, appmessage.CmdUnexpectedPruningPoint, appmessage.CmdPruningPointUTXOSetChunk,
			appmessage.CmdBlockHeaders, appmessage.CmdIBDBlockLocatorHighestHash, appmessage.CmdBlockWithTrustedData,
			appmessage.CmdDoneBlocksWithTrustedData, appmessage.CmdIBDBlockLocatorHighestHashNotFound,
			appmessage.CmdDonePruningPointUTXOSetChunks, appmessage.CmdIBDBlock, appmessage.CmdPruningPoints,
			appmessage.CmdPruningPointProof,
		},
			isStopping, errChan, func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRelayInvs(m.Context(), incomingRoute,
					outgoingRoute, peer)
			},
		),

		m.RegisterFlow("HandleRelayBlockRequests", router, []appmessage.MessageCommand{appmessage.CmdRequestRelayBlocks}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRelayBlockRequests(m.Context(), incomingRoute, outgoingRoute, peer)
			},
		),

		m.RegisterFlow("HandleRequestBlockLocator", router,
			[]appmessage.MessageCommand{appmessage.CmdRequestBlockLocator}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRequestBlockLocator(m.Context(), incomingRoute, outgoingRoute)
			},
		),

		m.RegisterFlow("HandleRequestHeaders", router,
			[]appmessage.MessageCommand{appmessage.CmdRequestHeaders, appmessage.CmdRequestNextHeaders}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRequestHeaders(m.Context(), incomingRoute, outgoingRoute, peer)
			},
		),

		m.RegisterFlow("HandleIBDBlockRequests", router,
			[]appmessage.MessageCommand{appmessage.CmdRequestIBDBlocks}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleIBDBlockRequests(m.Context(), incomingRoute, outgoingRoute)
			},
		),

		m.RegisterFlow("HandleRequestPruningPointUTXOSet", router,
			[]appmessage.MessageCommand{appmessage.CmdRequestPruningPointUTXOSet,
				appmessage.CmdRequestNextPruningPointUTXOSetChunk}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRequestPruningPointUTXOSet(m.Context(), incomingRoute, outgoingRoute)
			},
		),

		m.RegisterFlow("HandlePruningPointAndItsAnticoneRequests", router,
			[]appmessage.MessageCommand{appmessage.CmdRequestPruningPointAndItsAnticone}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandlePruningPointAndItsAnticoneRequests(m.Context(), incomingRoute, outgoingRoute, peer)
			},
		),

		m.RegisterFlow("HandleIBDBlockLocator", router,
			[]appmessage.MessageCommand{appmessage.CmdIBDBlockLocator}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleIBDBlockLocator(m.Context(), incomingRoute, outgoingRoute, peer)
			},
		),

		m.RegisterFlow("HandlePruningPointProofRequests", router,
			[]appmessage.MessageCommand{appmessage.CmdRequestPruningPointProof}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandlePruningPointProofRequests(m.Context(), incomingRoute, outgoingRoute, peer)
			},
		),
	}
}

func registerPingFlows(m protocolManager, router *routerpkg.Router, isStopping *uint32, errChan chan error) []*common.Flow {
	outgoingRoute := router.OutgoingRoute()

	return []*common.Flow{
		m.RegisterFlow("ReceivePings", router, []appmessage.MessageCommand{appmessage.CmdPing}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ping.ReceivePings(m.Context(), incomingRoute, outgoingRoute)
			},
		),

		m.RegisterFlow("SendPings", router, []appmessage.MessageCommand{appmessage.CmdPong}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ping.SendPings(m.Context(), incomingRoute, outgoingRoute, peer)
			},
		),
	}
}

func registerTransactionRelayFlow(m protocolManager, router *routerpkg.Router, isStopping *uint32, errChan chan error) []*common.Flow {
	outgoingRoute := router.OutgoingRoute()

	return []*common.Flow{
		m.RegisterFlowWithCapacity("HandleRelayedTransactions", 10_000, router,
			[]appmessage.MessageCommand{appmessage.CmdInvTransaction, appmessage.CmdTx, appmessage.CmdTransactionNotFound}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return transactionrelay.HandleRelayedTransactions(m.Context(), incomingRoute, outgoingRoute)
			},
		),
		m.RegisterFlow("HandleRequestTransactions", router,
			[]appmessage.MessageCommand{appmessage.CmdRequestTransactions}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return transactionrelay.HandleRequestedTransactions(m.Context(), incomingRoute, outgoingRoute)
			},
		),
	}
}

func registerRejectsFlow(m protocolManager, router *routerpkg.Router, isStopping *uint32, errChan chan error) []*common.Flow {
	outgoingRoute := router.OutgoingRoute()

	return []*common.Flow{
		m.RegisterFlow("HandleRejects", router,
			[]appmessage.MessageCommand{appmessage.CmdReject}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return rejects.HandleRejects(m.Context(), incomingRoute, outgoingRoute)
			},
		),
	}
}
