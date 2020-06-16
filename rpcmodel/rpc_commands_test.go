// Copyright (c) 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcmodel_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kaspanet/kaspad/util/pointers"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/rpcmodel"
)

// TestRPCServerCommands tests all of the kaspa rpc server commands marshal and unmarshal
// into valid results include handling of optional fields being omitted in the
// marshalled command, while optional fields with defaults have the default
// assigned on unmarshalled commands.
func TestRPCServerCommands(t *testing.T) {
	t.Parallel()

	testID := 1
	tests := []struct {
		name         string
		newCmd       func() (interface{}, error)
		staticCmd    func() interface{}
		marshalled   string
		unmarshalled interface{}
	}{
		{
			name: "addManualNode",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("addManualNode", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewAddManualNodeCmd("127.0.0.1", nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"addManualNode","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &rpcmodel.AddManualNodeCmd{Addr: "127.0.0.1", OneTry: pointers.Bool(false)},
		},
		{
			name: "createRawTransaction",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("createRawTransaction", `[{"txId":"123","vout":1}]`,
					`{"456":0.0123}`)
			},
			staticCmd: func() interface{} {
				txInputs := []rpcmodel.TransactionInput{
					{TxID: "123", Vout: 1},
				}
				amounts := map[string]float64{"456": .0123}
				return rpcmodel.NewCreateRawTransactionCmd(txInputs, amounts, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"createRawTransaction","params":[[{"txId":"123","vout":1}],{"456":0.0123}],"id":1}`,
			unmarshalled: &rpcmodel.CreateRawTransactionCmd{
				Inputs:  []rpcmodel.TransactionInput{{TxID: "123", Vout: 1}},
				Amounts: map[string]float64{"456": .0123},
			},
		},
		{
			name: "createRawTransaction optional",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("createRawTransaction", `[{"txId":"123","vout":1}]`,
					`{"456":0.0123}`, int64(12312333333))
			},
			staticCmd: func() interface{} {
				txInputs := []rpcmodel.TransactionInput{
					{TxID: "123", Vout: 1},
				}
				amounts := map[string]float64{"456": .0123}
				return rpcmodel.NewCreateRawTransactionCmd(txInputs, amounts, pointers.Uint64(12312333333))
			},
			marshalled: `{"jsonrpc":"1.0","method":"createRawTransaction","params":[[{"txId":"123","vout":1}],{"456":0.0123},12312333333],"id":1}`,
			unmarshalled: &rpcmodel.CreateRawTransactionCmd{
				Inputs:   []rpcmodel.TransactionInput{{TxID: "123", Vout: 1}},
				Amounts:  map[string]float64{"456": .0123},
				LockTime: pointers.Uint64(12312333333),
			},
		},

		{
			name: "decodeRawTransaction",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("decodeRawTransaction", "123")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewDecodeRawTransactionCmd("123")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"decodeRawTransaction","params":["123"],"id":1}`,
			unmarshalled: &rpcmodel.DecodeRawTransactionCmd{HexTx: "123"},
		},
		{
			name: "decodeScript",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("decodeScript", "00")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewDecodeScriptCmd("00")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"decodeScript","params":["00"],"id":1}`,
			unmarshalled: &rpcmodel.DecodeScriptCmd{HexScript: "00"},
		},
		{
			name: "getAllManualNodesInfo",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getAllManualNodesInfo")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetAllManualNodesInfoCmd(nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getAllManualNodesInfo","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetAllManualNodesInfoCmd{Details: pointers.Bool(true)},
		},
		{
			name: "getSelectedTipHash",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getSelectedTipHash")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetSelectedTipHashCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getSelectedTipHash","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetSelectedTipHashCmd{},
		},
		{
			name: "getBlock",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getBlock", "123")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetBlockCmd("123", nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123"],"id":1}`,
			unmarshalled: &rpcmodel.GetBlockCmd{
				Hash:      "123",
				Verbose:   pointers.Bool(true),
				VerboseTx: pointers.Bool(false),
			},
		},
		{
			name: "getBlock required optional1",
			newCmd: func() (interface{}, error) {
				// Intentionally use a source param that is
				// more pointers than the destination to
				// exercise that path.
				verbosePtr := pointers.Bool(true)
				return rpcmodel.NewCommand("getBlock", "123", &verbosePtr)
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetBlockCmd("123", pointers.Bool(true), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true],"id":1}`,
			unmarshalled: &rpcmodel.GetBlockCmd{
				Hash:      "123",
				Verbose:   pointers.Bool(true),
				VerboseTx: pointers.Bool(false),
			},
		},
		{
			name: "getBlock required optional2",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getBlock", "123", true, true)
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetBlockCmd("123", pointers.Bool(true), pointers.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true,true],"id":1}`,
			unmarshalled: &rpcmodel.GetBlockCmd{
				Hash:      "123",
				Verbose:   pointers.Bool(true),
				VerboseTx: pointers.Bool(true),
			},
		},
		{
			name: "getBlock required optional3",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getBlock", "123", true, true, "456")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetBlockCmd("123", pointers.Bool(true), pointers.Bool(true), pointers.String("456"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true,true,"456"],"id":1}`,
			unmarshalled: &rpcmodel.GetBlockCmd{
				Hash:       "123",
				Verbose:    pointers.Bool(true),
				VerboseTx:  pointers.Bool(true),
				Subnetwork: pointers.String("456"),
			},
		},
		{
			name: "getBlocks",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getBlocks", true, true, "123")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetBlocksCmd(true, true, pointers.String("123"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlocks","params":[true,true,"123"],"id":1}`,
			unmarshalled: &rpcmodel.GetBlocksCmd{
				IncludeRawBlockData:     true,
				IncludeVerboseBlockData: true,
				LowHash:                 pointers.String("123"),
			},
		},
		{
			name: "getBlockDagInfo",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getBlockDagInfo")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetBlockDAGInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockDagInfo","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetBlockDAGInfoCmd{},
		},
		{
			name: "getBlockCount",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getBlockCount")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetBlockCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockCount","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetBlockCountCmd{},
		},
		{
			name: "getBlockHeader",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getBlockHeader", "123")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetBlockHeaderCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockHeader","params":["123"],"id":1}`,
			unmarshalled: &rpcmodel.GetBlockHeaderCmd{
				Hash:    "123",
				Verbose: pointers.Bool(true),
			},
		},
		{
			name: "getBlockTemplate",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getBlockTemplate")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetBlockTemplateCmd(nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetBlockTemplateCmd{Request: nil},
		},
		{
			name: "getBlockTemplate optional - template request",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getBlockTemplate", `{"mode":"template","payAddress":"kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3"}`)
			},
			staticCmd: func() interface{} {
				template := rpcmodel.TemplateRequest{
					Mode:       "template",
					PayAddress: "kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3",
				}
				return rpcmodel.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","payAddress":"kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3"}],"id":1}`,
			unmarshalled: &rpcmodel.GetBlockTemplateCmd{
				Request: &rpcmodel.TemplateRequest{
					Mode:       "template",
					PayAddress: "kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3",
				},
			},
		},
		{
			name: "getBlockTemplate optional - template request with tweaks",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getBlockTemplate", `{"mode":"template","sigOpLimit":500,"massLimit":100000000,"maxVersion":1,"payAddress":"kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3"}`)
			},
			staticCmd: func() interface{} {
				template := rpcmodel.TemplateRequest{
					Mode:       "template",
					PayAddress: "kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3",
					SigOpLimit: 500,
					MassLimit:  100000000,
					MaxVersion: 1,
				}
				return rpcmodel.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","sigOpLimit":500,"massLimit":100000000,"maxVersion":1,"payAddress":"kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3"}],"id":1}`,
			unmarshalled: &rpcmodel.GetBlockTemplateCmd{
				Request: &rpcmodel.TemplateRequest{
					Mode:       "template",
					PayAddress: "kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3",
					SigOpLimit: int64(500),
					MassLimit:  int64(100000000),
					MaxVersion: 1,
				},
			},
		},
		{
			name: "getBlockTemplate optional - template request with tweaks 2",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getBlockTemplate", `{"mode":"template","payAddress":"kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3","sigOpLimit":true,"massLimit":100000000,"maxVersion":1}`)
			},
			staticCmd: func() interface{} {
				template := rpcmodel.TemplateRequest{
					Mode:       "template",
					PayAddress: "kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3",
					SigOpLimit: true,
					MassLimit:  100000000,
					MaxVersion: 1,
				}
				return rpcmodel.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","sigOpLimit":true,"massLimit":100000000,"maxVersion":1,"payAddress":"kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3"}],"id":1}`,
			unmarshalled: &rpcmodel.GetBlockTemplateCmd{
				Request: &rpcmodel.TemplateRequest{
					Mode:       "template",
					PayAddress: "kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3",
					SigOpLimit: true,
					MassLimit:  int64(100000000),
					MaxVersion: 1,
				},
			},
		},
		{
			name: "getChainFromBlock",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getChainFromBlock", true, "123")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetChainFromBlockCmd(true, pointers.String("123"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getChainFromBlock","params":[true,"123"],"id":1}`,
			unmarshalled: &rpcmodel.GetChainFromBlockCmd{
				IncludeBlocks: true,
				StartHash:     pointers.String("123"),
			},
		},
		{
			name: "getDagTips",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getDagTips")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetDAGTipsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getDagTips","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetDAGTipsCmd{},
		},
		{
			name: "getConnectionCount",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getConnectionCount")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetConnectionCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getConnectionCount","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetConnectionCountCmd{},
		},
		{
			name: "getDifficulty",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getDifficulty")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetDifficultyCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getDifficulty","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetDifficultyCmd{},
		},
		{
			name: "getInfo",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getInfo")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getInfo","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetInfoCmd{},
		},
		{
			name: "getManualNodeInfo",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getManualNodeInfo", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetManualNodeInfoCmd("127.0.0.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getManualNodeInfo","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &rpcmodel.GetManualNodeInfoCmd{
				Node:    "127.0.0.1",
				Details: pointers.Bool(true),
			},
		},
		{
			name: "getMempoolEntry",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getMempoolEntry", "txhash")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetMempoolEntryCmd("txhash")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getMempoolEntry","params":["txhash"],"id":1}`,
			unmarshalled: &rpcmodel.GetMempoolEntryCmd{
				TxID: "txhash",
			},
		},
		{
			name: "getMempoolInfo",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getMempoolInfo")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetMempoolInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getMempoolInfo","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetMempoolInfoCmd{},
		},
		{
			name: "getNetworkInfo",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getNetworkInfo")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetNetworkInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getNetworkInfo","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetNetworkInfoCmd{},
		},
		{
			name: "getNetTotals",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getNetTotals")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetNetTotalsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getNetTotals","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetNetTotalsCmd{},
		},
		{
			name: "getPeerInfo",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getPeerInfo")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetPeerInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getPeerInfo","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetPeerInfoCmd{},
		},
		{
			name: "getRawMempool",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getRawMempool")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetRawMempoolCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawMempool","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetRawMempoolCmd{
				Verbose: pointers.Bool(false),
			},
		},
		{
			name: "getRawMempool optional",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getRawMempool", false)
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetRawMempoolCmd(pointers.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawMempool","params":[false],"id":1}`,
			unmarshalled: &rpcmodel.GetRawMempoolCmd{
				Verbose: pointers.Bool(false),
			},
		},
		{
			name: "getSubnetwork",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getSubnetwork", "123")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetSubnetworkCmd("123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getSubnetwork","params":["123"],"id":1}`,
			unmarshalled: &rpcmodel.GetSubnetworkCmd{
				SubnetworkID: "123",
			},
		},
		{
			name: "getTxOut",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getTxOut", "123", 1)
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetTxOutCmd("123", 1, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTxOut","params":["123",1],"id":1}`,
			unmarshalled: &rpcmodel.GetTxOutCmd{
				TxID:           "123",
				Vout:           1,
				IncludeMempool: pointers.Bool(true),
			},
		},
		{
			name: "getTxOut optional",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getTxOut", "123", 1, true)
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetTxOutCmd("123", 1, pointers.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTxOut","params":["123",1,true],"id":1}`,
			unmarshalled: &rpcmodel.GetTxOutCmd{
				TxID:           "123",
				Vout:           1,
				IncludeMempool: pointers.Bool(true),
			},
		},
		{
			name: "getTxOutSetInfo",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getTxOutSetInfo")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetTxOutSetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getTxOutSetInfo","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetTxOutSetInfoCmd{},
		},
		{
			name: "help",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("help")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewHelpCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":[],"id":1}`,
			unmarshalled: &rpcmodel.HelpCmd{
				Command: nil,
			},
		},
		{
			name: "help optional",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("help", "getBlock")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewHelpCmd(pointers.String("getBlock"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":["getBlock"],"id":1}`,
			unmarshalled: &rpcmodel.HelpCmd{
				Command: pointers.String("getBlock"),
			},
		},
		{
			name: "ping",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("ping")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewPingCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"ping","params":[],"id":1}`,
			unmarshalled: &rpcmodel.PingCmd{},
		},
		{
			name: "removeManualNode",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("removeManualNode", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewRemoveManualNodeCmd("127.0.0.1")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"removeManualNode","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &rpcmodel.RemoveManualNodeCmd{Addr: "127.0.0.1"},
		},
		{
			name: "sendRawTransaction",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("sendRawTransaction", "1122")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewSendRawTransactionCmd("1122", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendRawTransaction","params":["1122"],"id":1}`,
			unmarshalled: &rpcmodel.SendRawTransactionCmd{
				HexTx:         "1122",
				AllowHighFees: pointers.Bool(false),
			},
		},
		{
			name: "sendRawTransaction optional",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("sendRawTransaction", "1122", false)
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewSendRawTransactionCmd("1122", pointers.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendRawTransaction","params":["1122",false],"id":1}`,
			unmarshalled: &rpcmodel.SendRawTransactionCmd{
				HexTx:         "1122",
				AllowHighFees: pointers.Bool(false),
			},
		},
		{
			name: "stop",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("stop")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewStopCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stop","params":[],"id":1}`,
			unmarshalled: &rpcmodel.StopCmd{},
		},
		{
			name: "submitBlock",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("submitBlock", "112233")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewSubmitBlockCmd("112233", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitBlock","params":["112233"],"id":1}`,
			unmarshalled: &rpcmodel.SubmitBlockCmd{
				HexBlock: "112233",
				Options:  nil,
			},
		},
		{
			name: "submitBlock optional",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("submitBlock", "112233", `{"workId":"12345"}`)
			},
			staticCmd: func() interface{} {
				options := rpcmodel.SubmitBlockOptions{
					WorkID: "12345",
				}
				return rpcmodel.NewSubmitBlockCmd("112233", &options)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitBlock","params":["112233",{"workId":"12345"}],"id":1}`,
			unmarshalled: &rpcmodel.SubmitBlockCmd{
				HexBlock: "112233",
				Options: &rpcmodel.SubmitBlockOptions{
					WorkID: "12345",
				},
			},
		},
		{
			name: "uptime",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("uptime")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewUptimeCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"uptime","params":[],"id":1}`,
			unmarshalled: &rpcmodel.UptimeCmd{},
		},
		{
			name: "validateAddress",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("validateAddress", "1Address")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewValidateAddressCmd("1Address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"validateAddress","params":["1Address"],"id":1}`,
			unmarshalled: &rpcmodel.ValidateAddressCmd{
				Address: "1Address",
			},
		},
		{
			name: "debugLevel",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("debugLevel", "trace")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewDebugLevelCmd("trace")
			},
			marshalled: `{"jsonrpc":"1.0","method":"debugLevel","params":["trace"],"id":1}`,
			unmarshalled: &rpcmodel.DebugLevelCmd{
				LevelSpec: "trace",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("node", rpcmodel.NRemove, "1.1.1.1")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewNodeCmd("remove", "1.1.1.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["remove","1.1.1.1"],"id":1}`,
			unmarshalled: &rpcmodel.NodeCmd{
				SubCmd: rpcmodel.NRemove,
				Target: "1.1.1.1",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("node", rpcmodel.NDisconnect, "1.1.1.1")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewNodeCmd("disconnect", "1.1.1.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["disconnect","1.1.1.1"],"id":1}`,
			unmarshalled: &rpcmodel.NodeCmd{
				SubCmd: rpcmodel.NDisconnect,
				Target: "1.1.1.1",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("node", rpcmodel.NConnect, "1.1.1.1", "perm")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewNodeCmd("connect", "1.1.1.1", pointers.String("perm"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["connect","1.1.1.1","perm"],"id":1}`,
			unmarshalled: &rpcmodel.NodeCmd{
				SubCmd:        rpcmodel.NConnect,
				Target:        "1.1.1.1",
				ConnectSubCmd: pointers.String("perm"),
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("node", rpcmodel.NConnect, "1.1.1.1", "temp")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewNodeCmd("connect", "1.1.1.1", pointers.String("temp"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["connect","1.1.1.1","temp"],"id":1}`,
			unmarshalled: &rpcmodel.NodeCmd{
				SubCmd:        rpcmodel.NConnect,
				Target:        "1.1.1.1",
				ConnectSubCmd: pointers.String("temp"),
			},
		},
		{
			name: "getSelectedTip",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getSelectedTip")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetSelectedTipCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getSelectedTip","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetSelectedTipCmd{
				Verbose:   pointers.Bool(true),
				VerboseTx: pointers.Bool(false),
			},
		},
		{
			name: "getCurrentNet",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getCurrentNet")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetCurrentNetCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getCurrentNet","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetCurrentNetCmd{},
		},
		{
			name: "getHeaders",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getHeaders", "", "")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetHeadersCmd(
					"",
					"",
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getHeaders","params":["",""],"id":1}`,
			unmarshalled: &rpcmodel.GetHeadersCmd{
				LowHash:  "",
				HighHash: "",
			},
		},
		{
			name: "getHeaders - with arguments",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getHeaders", "000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16", "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetHeadersCmd(
					"000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16",
					"000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7",
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getHeaders","params":["000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16","000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"],"id":1}`,
			unmarshalled: &rpcmodel.GetHeadersCmd{
				LowHash:  "000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16",
				HighHash: "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7",
			},
		},
		{
			name: "getTopHeaders",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getTopHeaders")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetTopHeadersCmd(
					nil,
				)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getTopHeaders","params":[],"id":1}`,
			unmarshalled: &rpcmodel.GetTopHeadersCmd{},
		},
		{
			name: "getTopHeaders - with high hash",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("getTopHeaders", "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewGetTopHeadersCmd(
					pointers.String("000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"),
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTopHeaders","params":["000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"],"id":1}`,
			unmarshalled: &rpcmodel.GetTopHeadersCmd{
				HighHash: pointers.String("000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"),
			},
		},
		{
			name: "version",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("version")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewVersionCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"version","params":[],"id":1}`,
			unmarshalled: &rpcmodel.VersionCmd{},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Marshal the command as created by the new static command
		// creation function.
		marshalled, err := rpcmodel.MarshalCommand(testID, test.staticCmd())
		if err != nil {
			t.Errorf("MarshalCommand #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !bytes.Equal(marshalled, []byte(test.marshalled)) {
			t.Errorf("Test #%d (%s) unexpected marshalled data - "+
				"got %s, want %s", i, test.name, marshalled,
				test.marshalled)
			t.Errorf("\n%s\n%s", marshalled, test.marshalled)
			continue
		}

		// Ensure the command is created without error via the generic
		// new command creation function.
		cmd, err := test.newCmd()
		if err != nil {
			t.Errorf("Test #%d (%s) unexpected NewCommand error: %v ",
				i, test.name, err)
		}

		// Marshal the command as created by the generic new command
		// creation function.
		marshalled, err = rpcmodel.MarshalCommand(testID, cmd)
		if err != nil {
			t.Errorf("MarshalCommand #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !bytes.Equal(marshalled, []byte(test.marshalled)) {
			t.Errorf("Test #%d (%s) unexpected marshalled data - "+
				"got %s, want %s", i, test.name, marshalled,
				test.marshalled)
			continue
		}

		var request rpcmodel.Request
		if err := json.Unmarshal(marshalled, &request); err != nil {
			t.Errorf("Test #%d (%s) unexpected error while "+
				"unmarshalling JSON-RPC request: %v", i,
				test.name, err)
			continue
		}

		cmd, err = rpcmodel.UnmarshalCommand(&request)
		if err != nil {
			t.Errorf("UnmarshalCommand #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !reflect.DeepEqual(cmd, test.unmarshalled) {
			t.Errorf("Test #%d (%s) unexpected unmarshalled command "+
				"- got %s, want %s", i, test.name,
				fmt.Sprintf("(%T) %+[1]v", cmd),
				fmt.Sprintf("(%T) %+[1]v\n", test.unmarshalled))
			continue
		}
	}
}

// TestRPCServerCommandErrors ensures any errors that occur in the command during
// custom mashal and unmarshal are as expected.
func TestRPCServerCommandErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		result     interface{}
		marshalled string
		err        error
	}{
		{
			name:       "template request with invalid type",
			result:     &rpcmodel.TemplateRequest{},
			marshalled: `{"mode":1}`,
			err:        &json.UnmarshalTypeError{},
		},
		{
			name:       "invalid template request sigoplimit field",
			result:     &rpcmodel.TemplateRequest{},
			marshalled: `{"sigoplimit":"invalid"}`,
			err:        rpcmodel.Error{ErrorCode: rpcmodel.ErrInvalidType},
		},
		{
			name:       "invalid template request masslimit field",
			result:     &rpcmodel.TemplateRequest{},
			marshalled: `{"masslimit":"invalid"}`,
			err:        rpcmodel.Error{ErrorCode: rpcmodel.ErrInvalidType},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		err := json.Unmarshal([]byte(test.marshalled), &test.result)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("Test #%d (%s) wrong error - got %T (%[2]v), "+
				"want %T", i, test.name, err, test.err)
			continue
		}

		var testErr rpcmodel.Error
		if errors.As(err, &testErr) {
			var gotRPCModelErr rpcmodel.Error
			errors.As(err, &gotRPCModelErr)
			gotErrorCode := gotRPCModelErr.ErrorCode
			if gotErrorCode != testErr.ErrorCode {
				t.Errorf("Test #%d (%s) mismatched error code "+
					"- got %v (%v), want %v", i, test.name,
					gotErrorCode, testErr, testErr.ErrorCode)
				continue
			}
		}
	}
}
