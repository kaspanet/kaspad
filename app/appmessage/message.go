// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"fmt"
	"time"
)

// MaxMessagePayload is the maximum bytes a message can be regardless of other
// individual limits imposed by messages themselves.
const MaxMessagePayload = 1024 * 1024 * 32 // 32MB

// MessageCommand is a number in the header of a message that represents its type.
type MessageCommand uint32

func (cmd MessageCommand) String() string {
	cmdString, ok := ProtocolMessageCommandToString[cmd]
	if !ok {
		cmdString, ok = RPCMessageCommandToString[cmd]
	}
	if !ok {
		cmdString = "unknown command"
	}
	return fmt.Sprintf("%s [code %d]", cmdString, uint8(cmd))
}

// Commands used in kaspa message headers which describe the type of message.
const (
	// protocol
	CmdVersion MessageCommand = iota
	CmdVerAck
	CmdRequestAddresses
	CmdAddresses
	CmdRequestHeaders
	CmdBlock
	CmdTx
	CmdPing
	CmdPong
	CmdRequestBlockLocator
	CmdBlockLocator
	CmdInvRelayBlock
	CmdRequestRelayBlocks
	CmdInvTransaction
	CmdRequestTransactions
	CmdIBDBlock
	CmdDoneHeaders
	CmdTransactionNotFound
	CmdReject
	CmdHeader
	CmdRequestNextHeaders
	CmdRequestIBDRootUTXOSetAndBlock
	CmdIBDRootUTXOSetAndBlock
	CmdRequestIBDBlocks
	CmdIBDRootNotFound

	// rpc
	CmdGetCurrentNetworkRequestMessage
	CmdGetCurrentNetworkResponseMessage
	CmdSubmitBlockRequestMessage
	CmdSubmitBlockResponseMessage
	CmdGetBlockTemplateRequestMessage
	CmdGetBlockTemplateResponseMessage
	CmdGetBlockTemplateTransactionMessage
	CmdNotifyBlockAddedRequestMessage
	CmdNotifyBlockAddedResponseMessage
	CmdBlockAddedNotificationMessage
	CmdGetPeerAddressesRequestMessage
	CmdGetPeerAddressesResponseMessage
	CmdGetSelectedTipHashRequestMessage
	CmdGetSelectedTipHashResponseMessage
	CmdGetMempoolEntryRequestMessage
	CmdGetMempoolEntryResponseMessage
	CmdGetConnectedPeerInfoRequestMessage
	CmdGetConnectedPeerInfoResponseMessage
	CmdAddPeerRequestMessage
	CmdAddPeerResponseMessage
	CmdSubmitTransactionRequestMessage
	CmdSubmitTransactionResponseMessage
	CmdNotifyChainChangedRequestMessage
	CmdNotifyChainChangedResponseMessage
	CmdChainChangedNotificationMessage
	CmdGetBlockRequestMessage
	CmdGetBlockResponseMessage
	CmdGetSubnetworkRequestMessage
	CmdGetSubnetworkResponseMessage
	CmdGetChainFromBlockRequestMessage
	CmdGetChainFromBlockResponseMessage
	CmdGetBlocksRequestMessage
	CmdGetBlocksResponseMessage
	CmdGetBlockCountRequestMessage
	CmdGetBlockCountResponseMessage
	CmdGetBlockDAGInfoRequestMessage
	CmdGetBlockDAGInfoResponseMessage
	CmdResolveFinalityConflictRequestMessage
	CmdResolveFinalityConflictResponseMessage
	CmdNotifyFinalityConflictsRequestMessage
	CmdNotifyFinalityConflictsResponseMessage
	CmdFinalityConflictNotificationMessage
	CmdFinalityConflictResolvedNotificationMessage
	CmdGetMempoolEntriesRequestMessage
	CmdGetMempoolEntriesResponseMessage
	CmdShutDownRequestMessage
	CmdShutDownResponseMessage
	CmdGetHeadersRequestMessage
	CmdGetHeadersResponseMessage
	CmdNotifyUTXOsChangedRequestMessage
	CmdNotifyUTXOsChangedResponseMessage
	CmdUTXOsChangedNotificationMessage
	CmdGetUTXOsByAddressesRequestMessage
	CmdGetUTXOsByAddressesResponseMessage
	CmdGetVirtualSelectedParentBlueScoreRequestMessage
	CmdGetVirtualSelectedParentBlueScoreResponseMessage
	CmdNotifyVirtualSelectedParentBlueScoreChangedRequestMessage
	CmdNotifyVirtualSelectedParentBlueScoreChangedResponseMessage
	CmdVirtualSelectedParentBlueScoreChangedNotificationMessage
)

// ProtocolMessageCommandToString maps all MessageCommands to their string representation
var ProtocolMessageCommandToString = map[MessageCommand]string{
	CmdVersion:                       "Version",
	CmdVerAck:                        "VerAck",
	CmdRequestAddresses:              "RequestAddresses",
	CmdAddresses:                     "Addresses",
	CmdRequestHeaders:                "RequestHeaders",
	CmdBlock:                         "Block",
	CmdTx:                            "Tx",
	CmdPing:                          "Ping",
	CmdPong:                          "Pong",
	CmdRequestBlockLocator:           "RequestBlockLocator",
	CmdBlockLocator:                  "BlockLocator",
	CmdInvRelayBlock:                 "InvRelayBlock",
	CmdRequestRelayBlocks:            "RequestRelayBlocks",
	CmdInvTransaction:                "InvTransaction",
	CmdRequestTransactions:           "RequestTransactions",
	CmdIBDBlock:                      "IBDBlock",
	CmdDoneHeaders:                   "DoneHeaders",
	CmdTransactionNotFound:           "TransactionNotFound",
	CmdReject:                        "Reject",
	CmdHeader:                        "Header",
	CmdRequestNextHeaders:            "RequestNextHeaders",
	CmdRequestIBDRootUTXOSetAndBlock: "RequestPruningUTXOSetAndBlock",
	CmdIBDRootUTXOSetAndBlock:        "IBDRootUTXOSetAndBlock",
	CmdRequestIBDBlocks:              "RequestIBDBlocks",
	CmdIBDRootNotFound:               "IBDRootNotFound",
}

// RPCMessageCommandToString maps all MessageCommands to their string representation
var RPCMessageCommandToString = map[MessageCommand]string{
	CmdGetCurrentNetworkRequestMessage:                            "GetCurrentNetworkRequest",
	CmdGetCurrentNetworkResponseMessage:                           "GetCurrentNetworkResponse",
	CmdSubmitBlockRequestMessage:                                  "SubmitBlockRequest",
	CmdSubmitBlockResponseMessage:                                 "SubmitBlockResponse",
	CmdGetBlockTemplateRequestMessage:                             "GetBlockTemplateRequest",
	CmdGetBlockTemplateResponseMessage:                            "GetBlockTemplateResponse",
	CmdGetBlockTemplateTransactionMessage:                         "CmdGetBlockTemplateTransaction",
	CmdNotifyBlockAddedRequestMessage:                             "NotifyBlockAddedRequest",
	CmdNotifyBlockAddedResponseMessage:                            "NotifyBlockAddedResponse",
	CmdBlockAddedNotificationMessage:                              "BlockAddedNotification",
	CmdGetPeerAddressesRequestMessage:                             "GetPeerAddressesRequest",
	CmdGetPeerAddressesResponseMessage:                            "GetPeerAddressesResponse",
	CmdGetSelectedTipHashRequestMessage:                           "GetSelectedTipHashRequest",
	CmdGetSelectedTipHashResponseMessage:                          "GetSelectedTipHashResponse",
	CmdGetMempoolEntryRequestMessage:                              "GetMempoolEntryRequest",
	CmdGetMempoolEntryResponseMessage:                             "GetMempoolEntryResponse",
	CmdGetConnectedPeerInfoRequestMessage:                         "GetConnectedPeerInfoRequest",
	CmdGetConnectedPeerInfoResponseMessage:                        "GetConnectedPeerInfoResponse",
	CmdAddPeerRequestMessage:                                      "AddPeerRequest",
	CmdAddPeerResponseMessage:                                     "AddPeerResponse",
	CmdSubmitTransactionRequestMessage:                            "SubmitTransactionRequest",
	CmdSubmitTransactionResponseMessage:                           "SubmitTransactionResponse",
	CmdNotifyChainChangedRequestMessage:                           "NotifyChainChangedRequest",
	CmdNotifyChainChangedResponseMessage:                          "NotifyChainChangedResponse",
	CmdChainChangedNotificationMessage:                            "ChainChangedNotification",
	CmdGetBlockRequestMessage:                                     "GetBlockRequest",
	CmdGetBlockResponseMessage:                                    "GetBlockResponse",
	CmdGetSubnetworkRequestMessage:                                "GetSubnetworkRequest",
	CmdGetSubnetworkResponseMessage:                               "GetSubnetworkResponse",
	CmdGetChainFromBlockRequestMessage:                            "GetChainFromBlockRequest",
	CmdGetChainFromBlockResponseMessage:                           "GetChainFromBlockResponse",
	CmdGetBlocksRequestMessage:                                    "GetBlocksRequest",
	CmdGetBlocksResponseMessage:                                   "GetBlocksResponse",
	CmdGetBlockCountRequestMessage:                                "GetBlockCountRequest",
	CmdGetBlockCountResponseMessage:                               "GetBlockCountResponse",
	CmdGetBlockDAGInfoRequestMessage:                              "GetBlockDAGInfoRequest",
	CmdGetBlockDAGInfoResponseMessage:                             "GetBlockDAGInfoResponse",
	CmdResolveFinalityConflictRequestMessage:                      "ResolveFinalityConflictRequest",
	CmdResolveFinalityConflictResponseMessage:                     "ResolveFinalityConflictResponse",
	CmdNotifyFinalityConflictsRequestMessage:                      "NotifyFinalityConflictsRequest",
	CmdNotifyFinalityConflictsResponseMessage:                     "NotifyFinalityConflictsResponse",
	CmdFinalityConflictNotificationMessage:                        "FinalityConflictNotification",
	CmdFinalityConflictResolvedNotificationMessage:                "FinalityConflictResolvedNotification",
	CmdGetMempoolEntriesRequestMessage:                            "GetMempoolEntriesRequest",
	CmdGetMempoolEntriesResponseMessage:                           "GetMempoolEntriesResponse",
	CmdGetHeadersRequestMessage:                                   "GetHeadersRequest",
	CmdGetHeadersResponseMessage:                                  "GetHeadersResponse",
	CmdNotifyUTXOsChangedRequestMessage:                           "NotifyUTXOsChangedRequest",
	CmdNotifyUTXOsChangedResponseMessage:                          "NotifyUTXOsChangedResponse",
	CmdUTXOsChangedNotificationMessage:                            "UTXOsChangedNotification",
	CmdGetUTXOsByAddressesRequestMessage:                          "GetUTXOsByAddressesRequest",
	CmdGetUTXOsByAddressesResponseMessage:                         "GetUTXOsByAddressesResponse",
	CmdGetVirtualSelectedParentBlueScoreRequestMessage:            "GetVirtualSelectedParentBlueScoreRequest",
	CmdGetVirtualSelectedParentBlueScoreResponseMessage:           "GetVirtualSelectedParentBlueScoreResponse",
	CmdNotifyVirtualSelectedParentBlueScoreChangedRequestMessage:  "NotifyVirtualSelectedParentBlueScoreChangedRequest",
	CmdNotifyVirtualSelectedParentBlueScoreChangedResponseMessage: "NotifyVirtualSelectedParentBlueScoreChangedResponse",
	CmdVirtualSelectedParentBlueScoreChangedNotificationMessage:   "VirtualSelectedParentBlueScoreChangedNotification",
}

// Message is an interface that describes a kaspa message. A type that
// implements Message has complete control over the representation of its data
// and may therefore contain additional or fewer fields than those which
// are used directly in the protocol encoded message.
type Message interface {
	Command() MessageCommand
	MessageNumber() uint64
	SetMessageNumber(index uint64)
	ReceivedAt() time.Time
	SetReceivedAt(receivedAt time.Time)
}
