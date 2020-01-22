// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// NOTE: This file is intended to house the RPC commands that are supported by
// a kaspa rpc server.

package rpcmodel

import (
	"encoding/json"
	"fmt"
)

// AddManualNodeCmd defines the addManualNode JSON-RPC command.
type AddManualNodeCmd struct {
	Addr   string
	OneTry *bool `jsonrpcdefault:"false"`
}

// NewAddManualNodeCmd returns a new instance which can be used to issue an addManualNode
// JSON-RPC command.
func NewAddManualNodeCmd(addr string, oneTry *bool) *AddManualNodeCmd {
	return &AddManualNodeCmd{
		Addr:   addr,
		OneTry: oneTry,
	}
}

// RemoveManualNodeCmd defines the removeManualNode JSON-RPC command.
type RemoveManualNodeCmd struct {
	Addr string
}

// NewRemoveManualNodeCmd returns a new instance which can be used to issue an removeManualNode
// JSON-RPC command.
func NewRemoveManualNodeCmd(addr string) *RemoveManualNodeCmd {
	return &RemoveManualNodeCmd{
		Addr: addr,
	}
}

// TransactionInput represents the inputs to a transaction. Specifically a
// transaction hash and output number pair.
type TransactionInput struct {
	TxID string `json:"txId"`
	Vout uint32 `json:"vout"`
}

// CreateRawTransactionCmd defines the createRawTransaction JSON-RPC command.
type CreateRawTransactionCmd struct {
	Inputs   []TransactionInput
	Amounts  map[string]float64 `jsonrpcusage:"{\"address\":amount,...}"` // In KAS
	LockTime *uint64
}

// NewCreateRawTransactionCmd returns a new instance which can be used to issue
// a createRawTransaction JSON-RPC command.
//
// Amounts are in KAS.
func NewCreateRawTransactionCmd(inputs []TransactionInput, amounts map[string]float64,
	lockTime *uint64) *CreateRawTransactionCmd {

	return &CreateRawTransactionCmd{
		Inputs:   inputs,
		Amounts:  amounts,
		LockTime: lockTime,
	}
}

// DecodeRawTransactionCmd defines the decodeRawTransaction JSON-RPC command.
type DecodeRawTransactionCmd struct {
	HexTx string
}

// NewDecodeRawTransactionCmd returns a new instance which can be used to issue
// a decodeRawTransaction JSON-RPC command.
func NewDecodeRawTransactionCmd(hexTx string) *DecodeRawTransactionCmd {
	return &DecodeRawTransactionCmd{
		HexTx: hexTx,
	}
}

// DecodeScriptCmd defines the decodeScript JSON-RPC command.
type DecodeScriptCmd struct {
	HexScript string
}

// NewDecodeScriptCmd returns a new instance which can be used to issue a
// decodeScript JSON-RPC command.
func NewDecodeScriptCmd(hexScript string) *DecodeScriptCmd {
	return &DecodeScriptCmd{
		HexScript: hexScript,
	}
}

// GetManualNodeInfoCmd defines the getManualNodeInfo JSON-RPC command.
type GetManualNodeInfoCmd struct {
	Node    string
	Details *bool `jsonrpcdefault:"true"`
}

// NewGetManualNodeInfoCmd returns a new instance which can be used to issue a
// getManualNodeInfo JSON-RPC command.
func NewGetManualNodeInfoCmd(node string, details *bool) *GetManualNodeInfoCmd {
	return &GetManualNodeInfoCmd{
		Details: details,
		Node:    node,
	}
}

// GetAllManualNodesInfoCmd defines the getAllManualNodesInfo JSON-RPC command.
type GetAllManualNodesInfoCmd struct {
	Details *bool `jsonrpcdefault:"true"`
}

// NewGetAllManualNodesInfoCmd returns a new instance which can be used to issue a
// getAllManualNodesInfo JSON-RPC command.
func NewGetAllManualNodesInfoCmd(details *bool) *GetAllManualNodesInfoCmd {
	return &GetAllManualNodesInfoCmd{
		Details: details,
	}
}

// GetSelectedTipHashCmd defines the getSelectedTipHash JSON-RPC command.
type GetSelectedTipHashCmd struct{}

// NewGetSelectedTipHashCmd returns a new instance which can be used to issue a
// getSelectedTipHash JSON-RPC command.
func NewGetSelectedTipHashCmd() *GetSelectedTipHashCmd {
	return &GetSelectedTipHashCmd{}
}

// GetBlockCmd defines the getBlock JSON-RPC command.
type GetBlockCmd struct {
	Hash       string
	Verbose    *bool `jsonrpcdefault:"true"`
	VerboseTx  *bool `jsonrpcdefault:"false"`
	Subnetwork *string
}

// NewGetBlockCmd returns a new instance which can be used to issue a getBlock
// JSON-RPC command.
//
// The parameters which are pointers indicate they are optional. Passing nil
// for optional parameters will use the default value.
func NewGetBlockCmd(hash string, verbose, verboseTx *bool, subnetworkID *string) *GetBlockCmd {
	return &GetBlockCmd{
		Hash:       hash,
		Verbose:    verbose,
		VerboseTx:  verboseTx,
		Subnetwork: subnetworkID,
	}
}

// GetBlocksCmd defines the getBlocks JSON-RPC command.
type GetBlocksCmd struct {
	IncludeRawBlockData     bool    `json:"includeRawBlockData"`
	IncludeVerboseBlockData bool    `json:"includeVerboseBlockData"`
	LowHash                 *string `json:"lowHash"`
}

// NewGetBlocksCmd returns a new instance which can be used to issue a
// GetGetBlocks JSON-RPC command.
func NewGetBlocksCmd(includeRawBlockData bool, includeVerboseBlockData bool, lowHash *string) *GetBlocksCmd {
	return &GetBlocksCmd{
		IncludeRawBlockData:     includeRawBlockData,
		IncludeVerboseBlockData: includeVerboseBlockData,
		LowHash:                 lowHash,
	}
}

// GetBlockDAGInfoCmd defines the getBlockDagInfo JSON-RPC command.
type GetBlockDAGInfoCmd struct{}

// NewGetBlockDAGInfoCmd returns a new instance which can be used to issue a
// getBlockDagInfo JSON-RPC command.
func NewGetBlockDAGInfoCmd() *GetBlockDAGInfoCmd {
	return &GetBlockDAGInfoCmd{}
}

// GetBlockCountCmd defines the getBlockCount JSON-RPC command.
type GetBlockCountCmd struct{}

// NewGetBlockCountCmd returns a new instance which can be used to issue a
// getBlockCount JSON-RPC command.
func NewGetBlockCountCmd() *GetBlockCountCmd {
	return &GetBlockCountCmd{}
}

// GetBlockHeaderCmd defines the getBlockHeader JSON-RPC command.
type GetBlockHeaderCmd struct {
	Hash    string
	Verbose *bool `jsonrpcdefault:"true"`
}

// NewGetBlockHeaderCmd returns a new instance which can be used to issue a
// getBlockHeader JSON-RPC command.
func NewGetBlockHeaderCmd(hash string, verbose *bool) *GetBlockHeaderCmd {
	return &GetBlockHeaderCmd{
		Hash:    hash,
		Verbose: verbose,
	}
}

// TemplateRequest is a request object as defined in BIP22. It is optionally
// provided as an pointer argument to GetBlockTemplateCmd.
type TemplateRequest struct {
	Mode         string   `json:"mode,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`

	// Optional long polling.
	LongPollID string `json:"longPollId,omitempty"`

	// Optional template tweaking. SigOpLimit and MassLimit can be int64
	// or bool.
	SigOpLimit interface{} `json:"sigOpLimit,omitempty"`
	MassLimit  interface{} `json:"massLimit,omitempty"`
	MaxVersion uint32      `json:"maxVersion,omitempty"`

	// Basic pool extension from BIP 0023.
	Target string `json:"target,omitempty"`

	// Block proposal from BIP 0023. Data is only provided when Mode is
	// "proposal".
	Data   string `json:"data,omitempty"`
	WorkID string `json:"workId,omitempty"`
}

// convertTemplateRequestField potentially converts the provided value as
// needed.
func convertTemplateRequestField(fieldName string, iface interface{}) (interface{}, error) {
	switch val := iface.(type) {
	case nil:
		return nil, nil
	case bool:
		return val, nil
	case float64:
		if val == float64(int64(val)) {
			return int64(val), nil
		}
	}

	str := fmt.Sprintf("the %s field must be unspecified, a boolean, or "+
		"a 64-bit integer", fieldName)
	return nil, makeError(ErrInvalidType, str)
}

// UnmarshalJSON provides a custom Unmarshal method for TemplateRequest. This
// is necessary because the SigOpLimit and MassLimit fields can only be specific
// types.
func (t *TemplateRequest) UnmarshalJSON(data []byte) error {
	type templateRequest TemplateRequest

	request := (*templateRequest)(t)
	if err := json.Unmarshal(data, &request); err != nil {
		return err
	}

	// The SigOpLimit field can only be nil, bool, or int64.
	val, err := convertTemplateRequestField("sigOpLimit", request.SigOpLimit)
	if err != nil {
		return err
	}
	request.SigOpLimit = val

	// The MassLimit field can only be nil, bool, or int64.
	val, err = convertTemplateRequestField("massLimit", request.MassLimit)
	if err != nil {
		return err
	}
	request.MassLimit = val

	return nil
}

// GetBlockTemplateCmd defines the getBlockTemplate JSON-RPC command.
type GetBlockTemplateCmd struct {
	Request *TemplateRequest
}

// NewGetBlockTemplateCmd returns a new instance which can be used to issue a
// getBlockTemplate JSON-RPC command.
//
// The parameters which are pointers indicate they are optional. Passing nil
// for optional parameters will use the default value.
func NewGetBlockTemplateCmd(request *TemplateRequest) *GetBlockTemplateCmd {
	return &GetBlockTemplateCmd{
		Request: request,
	}
}

// GetChainFromBlockCmd defines the getChainFromBlock JSON-RPC command.
type GetChainFromBlockCmd struct {
	IncludeBlocks bool    `json:"includeBlocks"`
	LowHash       *string `json:"lowHash"`
}

// NewGetChainFromBlockCmd returns a new instance which can be used to issue a
// GetChainFromBlock JSON-RPC command.
func NewGetChainFromBlockCmd(includeBlocks bool, lowHash *string) *GetChainFromBlockCmd {
	return &GetChainFromBlockCmd{
		IncludeBlocks: includeBlocks,
		LowHash:       lowHash,
	}
}

// GetDAGTipsCmd defines the getDagTips JSON-RPC command.
type GetDAGTipsCmd struct{}

// NewGetDAGTipsCmd returns a new instance which can be used to issue a
// getDagTips JSON-RPC command.
func NewGetDAGTipsCmd() *GetDAGTipsCmd {
	return &GetDAGTipsCmd{}
}

// GetConnectionCountCmd defines the getConnectionCount JSON-RPC command.
type GetConnectionCountCmd struct{}

// NewGetConnectionCountCmd returns a new instance which can be used to issue a
// getConnectionCount JSON-RPC command.
func NewGetConnectionCountCmd() *GetConnectionCountCmd {
	return &GetConnectionCountCmd{}
}

// GetDifficultyCmd defines the getDifficulty JSON-RPC command.
type GetDifficultyCmd struct{}

// NewGetDifficultyCmd returns a new instance which can be used to issue a
// getDifficulty JSON-RPC command.
func NewGetDifficultyCmd() *GetDifficultyCmd {
	return &GetDifficultyCmd{}
}

// GetInfoCmd defines the getInfo JSON-RPC command.
type GetInfoCmd struct{}

// NewGetInfoCmd returns a new instance which can be used to issue a
// getInfo JSON-RPC command.
func NewGetInfoCmd() *GetInfoCmd {
	return &GetInfoCmd{}
}

// GetMempoolEntryCmd defines the getMempoolEntry JSON-RPC command.
type GetMempoolEntryCmd struct {
	TxID string
}

// NewGetMempoolEntryCmd returns a new instance which can be used to issue a
// getMempoolEntry JSON-RPC command.
func NewGetMempoolEntryCmd(txHash string) *GetMempoolEntryCmd {
	return &GetMempoolEntryCmd{
		TxID: txHash,
	}
}

// GetMempoolInfoCmd defines the getMempoolInfo JSON-RPC command.
type GetMempoolInfoCmd struct{}

// NewGetMempoolInfoCmd returns a new instance which can be used to issue a
// getmempool JSON-RPC command.
func NewGetMempoolInfoCmd() *GetMempoolInfoCmd {
	return &GetMempoolInfoCmd{}
}

// GetNetworkInfoCmd defines the getNetworkInfo JSON-RPC command.
type GetNetworkInfoCmd struct{}

// NewGetNetworkInfoCmd returns a new instance which can be used to issue a
// getNetworkInfo JSON-RPC command.
func NewGetNetworkInfoCmd() *GetNetworkInfoCmd {
	return &GetNetworkInfoCmd{}
}

// GetNetTotalsCmd defines the getNetTotals JSON-RPC command.
type GetNetTotalsCmd struct{}

// NewGetNetTotalsCmd returns a new instance which can be used to issue a
// getNetTotals JSON-RPC command.
func NewGetNetTotalsCmd() *GetNetTotalsCmd {
	return &GetNetTotalsCmd{}
}

// GetPeerInfoCmd defines the getPeerInfo JSON-RPC command.
type GetPeerInfoCmd struct{}

// NewGetPeerInfoCmd returns a new instance which can be used to issue a getpeer
// JSON-RPC command.
func NewGetPeerInfoCmd() *GetPeerInfoCmd {
	return &GetPeerInfoCmd{}
}

// GetRawMempoolCmd defines the getmempool JSON-RPC command.
type GetRawMempoolCmd struct {
	Verbose *bool `jsonrpcdefault:"false"`
}

// NewGetRawMempoolCmd returns a new instance which can be used to issue a
// getRawMempool JSON-RPC command.
//
// The parameters which are pointers indicate they are optional. Passing nil
// for optional parameters will use the default value.
func NewGetRawMempoolCmd(verbose *bool) *GetRawMempoolCmd {
	return &GetRawMempoolCmd{
		Verbose: verbose,
	}
}

// GetRawTransactionCmd defines the getRawTransaction JSON-RPC command.
type GetRawTransactionCmd struct {
	TxID    string
	Verbose *int `jsonrpcdefault:"0"`
}

// NewGetRawTransactionCmd returns a new instance which can be used to issue a
// getRawTransaction JSON-RPC command.
//
// The parameters which are pointers indicate they are optional. Passing nil
// for optional parameters will use the default value.
func NewGetRawTransactionCmd(txID string, verbose *int) *GetRawTransactionCmd {
	return &GetRawTransactionCmd{
		TxID:    txID,
		Verbose: verbose,
	}
}

// GetSubnetworkCmd defines the getSubnetwork JSON-RPC command.
type GetSubnetworkCmd struct {
	SubnetworkID string
}

// NewGetSubnetworkCmd returns a new instance which can be used to issue a
// getSubnetworkCmd command.
func NewGetSubnetworkCmd(subnetworkID string) *GetSubnetworkCmd {
	return &GetSubnetworkCmd{
		SubnetworkID: subnetworkID,
	}
}

// GetTxOutCmd defines the getTxOut JSON-RPC command.
type GetTxOutCmd struct {
	TxID           string
	Vout           uint32
	IncludeMempool *bool `jsonrpcdefault:"true"`
}

// NewGetTxOutCmd returns a new instance which can be used to issue a getTxOut
// JSON-RPC command.
//
// The parameters which are pointers indicate they are optional. Passing nil
// for optional parameters will use the default value.
func NewGetTxOutCmd(txHash string, vout uint32, includeMempool *bool) *GetTxOutCmd {
	return &GetTxOutCmd{
		TxID:           txHash,
		Vout:           vout,
		IncludeMempool: includeMempool,
	}
}

// GetTxOutSetInfoCmd defines the getTxOutSetInfo JSON-RPC command.
type GetTxOutSetInfoCmd struct{}

// NewGetTxOutSetInfoCmd returns a new instance which can be used to issue a
// getTxOutSetInfo JSON-RPC command.
func NewGetTxOutSetInfoCmd() *GetTxOutSetInfoCmd {
	return &GetTxOutSetInfoCmd{}
}

// HelpCmd defines the help JSON-RPC command.
type HelpCmd struct {
	Command *string
}

// NewHelpCmd returns a new instance which can be used to issue a help JSON-RPC
// command.
//
// The parameters which are pointers indicate they are optional. Passing nil
// for optional parameters will use the default value.
func NewHelpCmd(command *string) *HelpCmd {
	return &HelpCmd{
		Command: command,
	}
}

// PingCmd defines the ping JSON-RPC command.
type PingCmd struct{}

// NewPingCmd returns a new instance which can be used to issue a ping JSON-RPC
// command.
func NewPingCmd() *PingCmd {
	return &PingCmd{}
}

// SearchRawTransactionsCmd defines the searchRawTransactions JSON-RPC command.
type SearchRawTransactionsCmd struct {
	Address     string
	Verbose     *bool `jsonrpcdefault:"true"`
	Skip        *int  `jsonrpcdefault:"0"`
	Count       *int  `jsonrpcdefault:"100"`
	VinExtra    *bool `jsonrpcdefault:"false"`
	Reverse     *bool `jsonrpcdefault:"false"`
	FilterAddrs *[]string
}

// NewSearchRawTransactionsCmd returns a new instance which can be used to issue a
// sendRawTransaction JSON-RPC command.
//
// The parameters which are pointers indicate they are optional. Passing nil
// for optional parameters will use the default value.
func NewSearchRawTransactionsCmd(address string, verbose *bool, skip, count *int, vinExtra, reverse *bool, filterAddrs *[]string) *SearchRawTransactionsCmd {
	return &SearchRawTransactionsCmd{
		Address:     address,
		Verbose:     verbose,
		Skip:        skip,
		Count:       count,
		VinExtra:    vinExtra,
		Reverse:     reverse,
		FilterAddrs: filterAddrs,
	}
}

// SendRawTransactionCmd defines the sendRawTransaction JSON-RPC command.
type SendRawTransactionCmd struct {
	HexTx         string
	AllowHighFees *bool `jsonrpcdefault:"false"`
}

// NewSendRawTransactionCmd returns a new instance which can be used to issue a
// sendRawTransaction JSON-RPC command.
//
// The parameters which are pointers indicate they are optional. Passing nil
// for optional parameters will use the default value.
func NewSendRawTransactionCmd(hexTx string, allowHighFees *bool) *SendRawTransactionCmd {
	return &SendRawTransactionCmd{
		HexTx:         hexTx,
		AllowHighFees: allowHighFees,
	}
}

// StopCmd defines the stop JSON-RPC command.
type StopCmd struct{}

// NewStopCmd returns a new instance which can be used to issue a stop JSON-RPC
// command.
func NewStopCmd() *StopCmd {
	return &StopCmd{}
}

// SubmitBlockOptions represents the optional options struct provided with a
// SubmitBlockCmd command.
type SubmitBlockOptions struct {
	// must be provided if server provided a workid with template.
	WorkID string `json:"workId,omitempty"`
}

// SubmitBlockCmd defines the submitBlock JSON-RPC command.
type SubmitBlockCmd struct {
	HexBlock string
	Options  *SubmitBlockOptions
}

// NewSubmitBlockCmd returns a new instance which can be used to issue a
// submitBlock JSON-RPC command.
//
// The parameters which are pointers indicate they are optional. Passing nil
// for optional parameters will use the default value.
func NewSubmitBlockCmd(hexBlock string, options *SubmitBlockOptions) *SubmitBlockCmd {
	return &SubmitBlockCmd{
		HexBlock: hexBlock,
		Options:  options,
	}
}

// UptimeCmd defines the uptime JSON-RPC command.
type UptimeCmd struct{}

// NewUptimeCmd returns a new instance which can be used to issue an uptime JSON-RPC command.
func NewUptimeCmd() *UptimeCmd {
	return &UptimeCmd{}
}

// ValidateAddressCmd defines the validateAddress JSON-RPC command.
type ValidateAddressCmd struct {
	Address string
}

// NewValidateAddressCmd returns a new instance which can be used to issue a
// validateAddress JSON-RPC command.
func NewValidateAddressCmd(address string) *ValidateAddressCmd {
	return &ValidateAddressCmd{
		Address: address,
	}
}

// NodeSubCmd defines the type used in the `node` JSON-RPC command for the
// sub command field.
type NodeSubCmd string

const (
	// NConnect indicates the specified host that should be connected to.
	NConnect NodeSubCmd = "connect"

	// NRemove indicates the specified peer that should be removed as a
	// persistent peer.
	NRemove NodeSubCmd = "remove"

	// NDisconnect indicates the specified peer should be disconnected.
	NDisconnect NodeSubCmd = "disconnect"
)

// NodeCmd defines the node JSON-RPC command.
type NodeCmd struct {
	SubCmd        NodeSubCmd `jsonrpcusage:"\"connect|remove|disconnect\""`
	Target        string
	ConnectSubCmd *string `jsonrpcusage:"\"perm|temp\""`
}

// NewNodeCmd returns a new instance which can be used to issue a `node`
// JSON-RPC command.
//
// The parameters which are pointers indicate they are optional. Passing nil
// for optional parameters will use the default value.
func NewNodeCmd(subCmd NodeSubCmd, target string, connectSubCmd *string) *NodeCmd {
	return &NodeCmd{
		SubCmd:        subCmd,
		Target:        target,
		ConnectSubCmd: connectSubCmd,
	}
}

// DebugLevelCmd defines the debugLevel JSON-RPC command.
type DebugLevelCmd struct {
	LevelSpec string
}

// NewDebugLevelCmd returns a new DebugLevelCmd which can be used to issue a
// debugLevel JSON-RPC command.
func NewDebugLevelCmd(levelSpec string) *DebugLevelCmd {
	return &DebugLevelCmd{
		LevelSpec: levelSpec,
	}
}

// GetSelectedTipCmd defines the getSelectedTip JSON-RPC command.
type GetSelectedTipCmd struct {
	Verbose   *bool `jsonrpcdefault:"true"`
	VerboseTx *bool `jsonrpcdefault:"false"`
}

// NewGetSelectedTipCmd returns a new instance which can be used to issue a
// getSelectedTip JSON-RPC command.
func NewGetSelectedTipCmd(verbose, verboseTx *bool) *GetSelectedTipCmd {
	return &GetSelectedTipCmd{
		Verbose:   verbose,
		VerboseTx: verboseTx,
	}
}

// GetCurrentNetCmd defines the getCurrentNet JSON-RPC command.
type GetCurrentNetCmd struct{}

// NewGetCurrentNetCmd returns a new instance which can be used to issue a
// getCurrentNet JSON-RPC command.
func NewGetCurrentNetCmd() *GetCurrentNetCmd {
	return &GetCurrentNetCmd{}
}

// GetTopHeadersCmd defined the getTopHeaders JSON-RPC command.
type GetTopHeadersCmd struct {
	HighHash *string `json:"highHash"`
}

// NewGetTopHeadersCmd returns a new instance which can be used to issue a
// getTopHeaders JSON-RPC command.
func NewGetTopHeadersCmd(highHash *string) *GetTopHeadersCmd {
	return &GetTopHeadersCmd{
		HighHash: highHash,
	}
}

// GetHeadersCmd defines the getHeaders JSON-RPC command.
type GetHeadersCmd struct {
	LowHash  string `json:"lowHash"`
	HighHash string `json:"highHash"`
}

// NewGetHeadersCmd returns a new instance which can be used to issue a
// getHeaders JSON-RPC command.
func NewGetHeadersCmd(lowHash, highHash string) *GetHeadersCmd {
	return &GetHeadersCmd{
		LowHash:  lowHash,
		HighHash: highHash,
	}
}

// VersionCmd defines the version JSON-RPC command.
type VersionCmd struct{}

// NewVersionCmd returns a new instance which can be used to issue a JSON-RPC
// version command.
func NewVersionCmd() *VersionCmd { return new(VersionCmd) }

func init() {
	// No special flags for commands in this file.
	flags := UsageFlag(0)

	MustRegisterCommand("addManualNode", (*AddManualNodeCmd)(nil), flags)
	MustRegisterCommand("createRawTransaction", (*CreateRawTransactionCmd)(nil), flags)
	MustRegisterCommand("decodeRawTransaction", (*DecodeRawTransactionCmd)(nil), flags)
	MustRegisterCommand("decodeScript", (*DecodeScriptCmd)(nil), flags)
	MustRegisterCommand("getAllManualNodesInfo", (*GetAllManualNodesInfoCmd)(nil), flags)
	MustRegisterCommand("getSelectedTipHash", (*GetSelectedTipHashCmd)(nil), flags)
	MustRegisterCommand("getBlock", (*GetBlockCmd)(nil), flags)
	MustRegisterCommand("getBlocks", (*GetBlocksCmd)(nil), flags)
	MustRegisterCommand("getBlockDagInfo", (*GetBlockDAGInfoCmd)(nil), flags)
	MustRegisterCommand("getBlockCount", (*GetBlockCountCmd)(nil), flags)
	MustRegisterCommand("getBlockHeader", (*GetBlockHeaderCmd)(nil), flags)
	MustRegisterCommand("getBlockTemplate", (*GetBlockTemplateCmd)(nil), flags)
	MustRegisterCommand("getChainFromBlock", (*GetChainFromBlockCmd)(nil), flags)
	MustRegisterCommand("getDagTips", (*GetDAGTipsCmd)(nil), flags)
	MustRegisterCommand("getConnectionCount", (*GetConnectionCountCmd)(nil), flags)
	MustRegisterCommand("getDifficulty", (*GetDifficultyCmd)(nil), flags)
	MustRegisterCommand("getInfo", (*GetInfoCmd)(nil), flags)
	MustRegisterCommand("getManualNodeInfo", (*GetManualNodeInfoCmd)(nil), flags)
	MustRegisterCommand("getMempoolEntry", (*GetMempoolEntryCmd)(nil), flags)
	MustRegisterCommand("getMempoolInfo", (*GetMempoolInfoCmd)(nil), flags)
	MustRegisterCommand("getNetworkInfo", (*GetNetworkInfoCmd)(nil), flags)
	MustRegisterCommand("getNetTotals", (*GetNetTotalsCmd)(nil), flags)
	MustRegisterCommand("getPeerInfo", (*GetPeerInfoCmd)(nil), flags)
	MustRegisterCommand("getRawMempool", (*GetRawMempoolCmd)(nil), flags)
	MustRegisterCommand("getRawTransaction", (*GetRawTransactionCmd)(nil), flags)
	MustRegisterCommand("getSubnetwork", (*GetSubnetworkCmd)(nil), flags)
	MustRegisterCommand("getTxOut", (*GetTxOutCmd)(nil), flags)
	MustRegisterCommand("getTxOutSetInfo", (*GetTxOutSetInfoCmd)(nil), flags)
	MustRegisterCommand("help", (*HelpCmd)(nil), flags)
	MustRegisterCommand("ping", (*PingCmd)(nil), flags)
	MustRegisterCommand("removeManualNode", (*RemoveManualNodeCmd)(nil), flags)
	MustRegisterCommand("searchRawTransactions", (*SearchRawTransactionsCmd)(nil), flags)
	MustRegisterCommand("sendRawTransaction", (*SendRawTransactionCmd)(nil), flags)
	MustRegisterCommand("stop", (*StopCmd)(nil), flags)
	MustRegisterCommand("submitBlock", (*SubmitBlockCmd)(nil), flags)
	MustRegisterCommand("uptime", (*UptimeCmd)(nil), flags)
	MustRegisterCommand("validateAddress", (*ValidateAddressCmd)(nil), flags)
	MustRegisterCommand("debugLevel", (*DebugLevelCmd)(nil), flags)
	MustRegisterCommand("node", (*NodeCmd)(nil), flags)
	MustRegisterCommand("getSelectedTip", (*GetSelectedTipCmd)(nil), flags)
	MustRegisterCommand("getCurrentNet", (*GetCurrentNetCmd)(nil), flags)
	MustRegisterCommand("getHeaders", (*GetHeadersCmd)(nil), flags)
	MustRegisterCommand("getTopHeaders", (*GetTopHeadersCmd)(nil), flags)
	MustRegisterCommand("version", (*VersionCmd)(nil), flags)
}
