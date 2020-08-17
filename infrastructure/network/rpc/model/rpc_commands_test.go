// Copyright (c) 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package model_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kaspanet/kaspad/util/pointers"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/infrastructure/network/rpc/model"
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
			name: "connect",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("connect", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return model.NewConnectCmd("127.0.0.1", nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"connect","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &model.ConnectCmd{Address: "127.0.0.1", IsPermanent: pointers.Bool(false)},
		},
		{
			name: "getSelectedTipHash",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getSelectedTipHash")
			},
			staticCmd: func() interface{} {
				return model.NewGetSelectedTipHashCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getSelectedTipHash","params":[],"id":1}`,
			unmarshalled: &model.GetSelectedTipHashCmd{},
		},
		{
			name: "getBlock",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getBlock", "123")
			},
			staticCmd: func() interface{} {
				return model.NewGetBlockCmd("123", nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123"],"id":1}`,
			unmarshalled: &model.GetBlockCmd{
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
				return model.NewCommand("getBlock", "123", &verbosePtr)
			},
			staticCmd: func() interface{} {
				return model.NewGetBlockCmd("123", pointers.Bool(true), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true],"id":1}`,
			unmarshalled: &model.GetBlockCmd{
				Hash:      "123",
				Verbose:   pointers.Bool(true),
				VerboseTx: pointers.Bool(false),
			},
		},
		{
			name: "getBlock required optional2",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getBlock", "123", true, true)
			},
			staticCmd: func() interface{} {
				return model.NewGetBlockCmd("123", pointers.Bool(true), pointers.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true,true],"id":1}`,
			unmarshalled: &model.GetBlockCmd{
				Hash:      "123",
				Verbose:   pointers.Bool(true),
				VerboseTx: pointers.Bool(true),
			},
		},
		{
			name: "getBlock required optional3",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getBlock", "123", true, true, "456")
			},
			staticCmd: func() interface{} {
				return model.NewGetBlockCmd("123", pointers.Bool(true), pointers.Bool(true), pointers.String("456"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true,true,"456"],"id":1}`,
			unmarshalled: &model.GetBlockCmd{
				Hash:       "123",
				Verbose:    pointers.Bool(true),
				VerboseTx:  pointers.Bool(true),
				Subnetwork: pointers.String("456"),
			},
		},
		{
			name: "getBlocks",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getBlocks", true, true, "123")
			},
			staticCmd: func() interface{} {
				return model.NewGetBlocksCmd(true, true, pointers.String("123"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlocks","params":[true,true,"123"],"id":1}`,
			unmarshalled: &model.GetBlocksCmd{
				IncludeRawBlockData:     true,
				IncludeVerboseBlockData: true,
				LowHash:                 pointers.String("123"),
			},
		},
		{
			name: "getBlockDagInfo",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getBlockDagInfo")
			},
			staticCmd: func() interface{} {
				return model.NewGetBlockDAGInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockDagInfo","params":[],"id":1}`,
			unmarshalled: &model.GetBlockDAGInfoCmd{},
		},
		{
			name: "getBlockCount",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getBlockCount")
			},
			staticCmd: func() interface{} {
				return model.NewGetBlockCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockCount","params":[],"id":1}`,
			unmarshalled: &model.GetBlockCountCmd{},
		},
		{
			name: "getBlockHeader",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getBlockHeader", "123")
			},
			staticCmd: func() interface{} {
				return model.NewGetBlockHeaderCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockHeader","params":["123"],"id":1}`,
			unmarshalled: &model.GetBlockHeaderCmd{
				Hash:    "123",
				Verbose: pointers.Bool(true),
			},
		},
		{
			name: "getBlockTemplate",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getBlockTemplate")
			},
			staticCmd: func() interface{} {
				return model.NewGetBlockTemplateCmd(nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[],"id":1}`,
			unmarshalled: &model.GetBlockTemplateCmd{Request: nil},
		},
		{
			name: "getBlockTemplate optional - template request",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getBlockTemplate", `{"mode":"template","payAddress":"kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3"}`)
			},
			staticCmd: func() interface{} {
				template := model.TemplateRequest{
					Mode:       "template",
					PayAddress: "kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3",
				}
				return model.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","payAddress":"kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3"}],"id":1}`,
			unmarshalled: &model.GetBlockTemplateCmd{
				Request: &model.TemplateRequest{
					Mode:       "template",
					PayAddress: "kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3",
				},
			},
		},
		{
			name: "getBlockTemplate optional - template request with tweaks",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getBlockTemplate", `{"mode":"template","sigOpLimit":500,"massLimit":100000000,"maxVersion":1,"payAddress":"kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3"}`)
			},
			staticCmd: func() interface{} {
				template := model.TemplateRequest{
					Mode:       "template",
					PayAddress: "kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3",
					SigOpLimit: 500,
					MassLimit:  100000000,
					MaxVersion: 1,
				}
				return model.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","sigOpLimit":500,"massLimit":100000000,"maxVersion":1,"payAddress":"kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3"}],"id":1}`,
			unmarshalled: &model.GetBlockTemplateCmd{
				Request: &model.TemplateRequest{
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
				return model.NewCommand("getBlockTemplate", `{"mode":"template","payAddress":"kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3","sigOpLimit":true,"massLimit":100000000,"maxVersion":1}`)
			},
			staticCmd: func() interface{} {
				template := model.TemplateRequest{
					Mode:       "template",
					PayAddress: "kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3",
					SigOpLimit: true,
					MassLimit:  100000000,
					MaxVersion: 1,
				}
				return model.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","sigOpLimit":true,"massLimit":100000000,"maxVersion":1,"payAddress":"kaspa:qph364lxa0ul5h0jrvl3u7xu8erc7mu3dv7prcn7x3"}],"id":1}`,
			unmarshalled: &model.GetBlockTemplateCmd{
				Request: &model.TemplateRequest{
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
				return model.NewCommand("getChainFromBlock", true, "123")
			},
			staticCmd: func() interface{} {
				return model.NewGetChainFromBlockCmd(true, pointers.String("123"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getChainFromBlock","params":[true,"123"],"id":1}`,
			unmarshalled: &model.GetChainFromBlockCmd{
				IncludeBlocks: true,
				StartHash:     pointers.String("123"),
			},
		},
		{
			name: "getDagTips",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getDagTips")
			},
			staticCmd: func() interface{} {
				return model.NewGetDAGTipsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getDagTips","params":[],"id":1}`,
			unmarshalled: &model.GetDAGTipsCmd{},
		},
		{
			name: "getConnectionCount",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getConnectionCount")
			},
			staticCmd: func() interface{} {
				return model.NewGetConnectionCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getConnectionCount","params":[],"id":1}`,
			unmarshalled: &model.GetConnectionCountCmd{},
		},
		{
			name: "getDifficulty",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getDifficulty")
			},
			staticCmd: func() interface{} {
				return model.NewGetDifficultyCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getDifficulty","params":[],"id":1}`,
			unmarshalled: &model.GetDifficultyCmd{},
		},
		{
			name: "getInfo",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getInfo")
			},
			staticCmd: func() interface{} {
				return model.NewGetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getInfo","params":[],"id":1}`,
			unmarshalled: &model.GetInfoCmd{},
		},
		{
			name: "getMempoolEntry",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getMempoolEntry", "txhash")
			},
			staticCmd: func() interface{} {
				return model.NewGetMempoolEntryCmd("txhash")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getMempoolEntry","params":["txhash"],"id":1}`,
			unmarshalled: &model.GetMempoolEntryCmd{
				TxID: "txhash",
			},
		},
		{
			name: "getMempoolInfo",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getMempoolInfo")
			},
			staticCmd: func() interface{} {
				return model.NewGetMempoolInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getMempoolInfo","params":[],"id":1}`,
			unmarshalled: &model.GetMempoolInfoCmd{},
		},
		{
			name: "getNetworkInfo",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getNetworkInfo")
			},
			staticCmd: func() interface{} {
				return model.NewGetNetworkInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getNetworkInfo","params":[],"id":1}`,
			unmarshalled: &model.GetNetworkInfoCmd{},
		},
		{
			name: "getNetTotals",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getNetTotals")
			},
			staticCmd: func() interface{} {
				return model.NewGetNetTotalsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getNetTotals","params":[],"id":1}`,
			unmarshalled: &model.GetNetTotalsCmd{},
		},
		{
			name: "getConnectedPeerInfo",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getConnectedPeerInfo")
			},
			staticCmd: func() interface{} {
				return model.NewGetConnectedPeerInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getConnectedPeerInfo","params":[],"id":1}`,
			unmarshalled: &model.GetConnectedPeerInfoCmd{},
		},
		{
			name: "getRawMempool",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getRawMempool")
			},
			staticCmd: func() interface{} {
				return model.NewGetRawMempoolCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawMempool","params":[],"id":1}`,
			unmarshalled: &model.GetRawMempoolCmd{
				Verbose: pointers.Bool(false),
			},
		},
		{
			name: "getRawMempool optional",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getRawMempool", false)
			},
			staticCmd: func() interface{} {
				return model.NewGetRawMempoolCmd(pointers.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawMempool","params":[false],"id":1}`,
			unmarshalled: &model.GetRawMempoolCmd{
				Verbose: pointers.Bool(false),
			},
		},
		{
			name: "getSubnetwork",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getSubnetwork", "123")
			},
			staticCmd: func() interface{} {
				return model.NewGetSubnetworkCmd("123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getSubnetwork","params":["123"],"id":1}`,
			unmarshalled: &model.GetSubnetworkCmd{
				SubnetworkID: "123",
			},
		},
		{
			name: "getTxOut",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getTxOut", "123", 1)
			},
			staticCmd: func() interface{} {
				return model.NewGetTxOutCmd("123", 1, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTxOut","params":["123",1],"id":1}`,
			unmarshalled: &model.GetTxOutCmd{
				TxID:           "123",
				Vout:           1,
				IncludeMempool: pointers.Bool(true),
			},
		},
		{
			name: "getTxOut optional",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getTxOut", "123", 1, true)
			},
			staticCmd: func() interface{} {
				return model.NewGetTxOutCmd("123", 1, pointers.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTxOut","params":["123",1,true],"id":1}`,
			unmarshalled: &model.GetTxOutCmd{
				TxID:           "123",
				Vout:           1,
				IncludeMempool: pointers.Bool(true),
			},
		},
		{
			name: "getTxOutSetInfo",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getTxOutSetInfo")
			},
			staticCmd: func() interface{} {
				return model.NewGetTxOutSetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getTxOutSetInfo","params":[],"id":1}`,
			unmarshalled: &model.GetTxOutSetInfoCmd{},
		},
		{
			name: "help",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("help")
			},
			staticCmd: func() interface{} {
				return model.NewHelpCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":[],"id":1}`,
			unmarshalled: &model.HelpCmd{
				Command: nil,
			},
		},
		{
			name: "help optional",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("help", "getBlock")
			},
			staticCmd: func() interface{} {
				return model.NewHelpCmd(pointers.String("getBlock"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":["getBlock"],"id":1}`,
			unmarshalled: &model.HelpCmd{
				Command: pointers.String("getBlock"),
			},
		},
		{
			name: "ping",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("ping")
			},
			staticCmd: func() interface{} {
				return model.NewPingCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"ping","params":[],"id":1}`,
			unmarshalled: &model.PingCmd{},
		},
		{
			name: "disconnect",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("disconnect", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return model.NewDisconnectCmd("127.0.0.1")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"disconnect","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &model.DisconnectCmd{Address: "127.0.0.1"},
		},
		{
			name: "sendRawTransaction",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("sendRawTransaction", "1122")
			},
			staticCmd: func() interface{} {
				return model.NewSendRawTransactionCmd("1122", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendRawTransaction","params":["1122"],"id":1}`,
			unmarshalled: &model.SendRawTransactionCmd{
				HexTx:         "1122",
				AllowHighFees: pointers.Bool(false),
			},
		},
		{
			name: "sendRawTransaction optional",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("sendRawTransaction", "1122", false)
			},
			staticCmd: func() interface{} {
				return model.NewSendRawTransactionCmd("1122", pointers.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendRawTransaction","params":["1122",false],"id":1}`,
			unmarshalled: &model.SendRawTransactionCmd{
				HexTx:         "1122",
				AllowHighFees: pointers.Bool(false),
			},
		},
		{
			name: "stop",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("stop")
			},
			staticCmd: func() interface{} {
				return model.NewStopCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stop","params":[],"id":1}`,
			unmarshalled: &model.StopCmd{},
		},
		{
			name: "submitBlock",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("submitBlock", "112233")
			},
			staticCmd: func() interface{} {
				return model.NewSubmitBlockCmd("112233", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitBlock","params":["112233"],"id":1}`,
			unmarshalled: &model.SubmitBlockCmd{
				HexBlock: "112233",
				Options:  nil,
			},
		},
		{
			name: "submitBlock optional",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("submitBlock", "112233", `{"workId":"12345"}`)
			},
			staticCmd: func() interface{} {
				options := model.SubmitBlockOptions{
					WorkID: "12345",
				}
				return model.NewSubmitBlockCmd("112233", &options)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitBlock","params":["112233",{"workId":"12345"}],"id":1}`,
			unmarshalled: &model.SubmitBlockCmd{
				HexBlock: "112233",
				Options: &model.SubmitBlockOptions{
					WorkID: "12345",
				},
			},
		},
		{
			name: "uptime",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("uptime")
			},
			staticCmd: func() interface{} {
				return model.NewUptimeCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"uptime","params":[],"id":1}`,
			unmarshalled: &model.UptimeCmd{},
		},
		{
			name: "validateAddress",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("validateAddress", "1Address")
			},
			staticCmd: func() interface{} {
				return model.NewValidateAddressCmd("1Address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"validateAddress","params":["1Address"],"id":1}`,
			unmarshalled: &model.ValidateAddressCmd{
				Address: "1Address",
			},
		},
		{
			name: "debugLevel",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("debugLevel", "trace")
			},
			staticCmd: func() interface{} {
				return model.NewDebugLevelCmd("trace")
			},
			marshalled: `{"jsonrpc":"1.0","method":"debugLevel","params":["trace"],"id":1}`,
			unmarshalled: &model.DebugLevelCmd{
				LevelSpec: "trace",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("node", model.NRemove, "1.1.1.1")
			},
			staticCmd: func() interface{} {
				return model.NewNodeCmd("remove", "1.1.1.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["remove","1.1.1.1"],"id":1}`,
			unmarshalled: &model.NodeCmd{
				SubCmd: model.NRemove,
				Target: "1.1.1.1",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("node", model.NDisconnect, "1.1.1.1")
			},
			staticCmd: func() interface{} {
				return model.NewNodeCmd("disconnect", "1.1.1.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["disconnect","1.1.1.1"],"id":1}`,
			unmarshalled: &model.NodeCmd{
				SubCmd: model.NDisconnect,
				Target: "1.1.1.1",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("node", model.NConnect, "1.1.1.1", "perm")
			},
			staticCmd: func() interface{} {
				return model.NewNodeCmd("connect", "1.1.1.1", pointers.String("perm"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["connect","1.1.1.1","perm"],"id":1}`,
			unmarshalled: &model.NodeCmd{
				SubCmd:        model.NConnect,
				Target:        "1.1.1.1",
				ConnectSubCmd: pointers.String("perm"),
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("node", model.NConnect, "1.1.1.1", "temp")
			},
			staticCmd: func() interface{} {
				return model.NewNodeCmd("connect", "1.1.1.1", pointers.String("temp"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["connect","1.1.1.1","temp"],"id":1}`,
			unmarshalled: &model.NodeCmd{
				SubCmd:        model.NConnect,
				Target:        "1.1.1.1",
				ConnectSubCmd: pointers.String("temp"),
			},
		},
		{
			name: "getSelectedTip",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getSelectedTip")
			},
			staticCmd: func() interface{} {
				return model.NewGetSelectedTipCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getSelectedTip","params":[],"id":1}`,
			unmarshalled: &model.GetSelectedTipCmd{
				Verbose:   pointers.Bool(true),
				VerboseTx: pointers.Bool(false),
			},
		},
		{
			name: "getCurrentNet",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getCurrentNet")
			},
			staticCmd: func() interface{} {
				return model.NewGetCurrentNetCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getCurrentNet","params":[],"id":1}`,
			unmarshalled: &model.GetCurrentNetCmd{},
		},
		{
			name: "getHeaders",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getHeaders", "", "")
			},
			staticCmd: func() interface{} {
				return model.NewGetHeadersCmd(
					"",
					"",
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getHeaders","params":["",""],"id":1}`,
			unmarshalled: &model.GetHeadersCmd{
				LowHash:  "",
				HighHash: "",
			},
		},
		{
			name: "getHeaders - with arguments",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getHeaders", "000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16", "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7")
			},
			staticCmd: func() interface{} {
				return model.NewGetHeadersCmd(
					"000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16",
					"000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7",
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getHeaders","params":["000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16","000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"],"id":1}`,
			unmarshalled: &model.GetHeadersCmd{
				LowHash:  "000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16",
				HighHash: "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7",
			},
		},
		{
			name: "getTopHeaders",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getTopHeaders")
			},
			staticCmd: func() interface{} {
				return model.NewGetTopHeadersCmd(
					nil,
				)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getTopHeaders","params":[],"id":1}`,
			unmarshalled: &model.GetTopHeadersCmd{},
		},
		{
			name: "getTopHeaders - with high hash",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("getTopHeaders", "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7")
			},
			staticCmd: func() interface{} {
				return model.NewGetTopHeadersCmd(
					pointers.String("000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"),
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTopHeaders","params":["000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"],"id":1}`,
			unmarshalled: &model.GetTopHeadersCmd{
				HighHash: pointers.String("000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"),
			},
		},
		{
			name: "version",
			newCmd: func() (interface{}, error) {
				return model.NewCommand("version")
			},
			staticCmd: func() interface{} {
				return model.NewVersionCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"version","params":[],"id":1}`,
			unmarshalled: &model.VersionCmd{},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Marshal the command as created by the new static command
		// creation function.
		marshalled, err := model.MarshalCommand(testID, test.staticCmd())
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
		marshalled, err = model.MarshalCommand(testID, cmd)
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

		var request model.Request
		if err := json.Unmarshal(marshalled, &request); err != nil {
			t.Errorf("Test #%d (%s) unexpected error while "+
				"unmarshalling JSON-RPC request: %v", i,
				test.name, err)
			continue
		}

		cmd, err = model.UnmarshalCommand(&request)
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
			result:     &model.TemplateRequest{},
			marshalled: `{"mode":1}`,
			err:        &json.UnmarshalTypeError{},
		},
		{
			name:       "invalid template request sigoplimit field",
			result:     &model.TemplateRequest{},
			marshalled: `{"sigoplimit":"invalid"}`,
			err:        model.Error{ErrorCode: model.ErrInvalidType},
		},
		{
			name:       "invalid template request masslimit field",
			result:     &model.TemplateRequest{},
			marshalled: `{"masslimit":"invalid"}`,
			err:        model.Error{ErrorCode: model.ErrInvalidType},
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

		var testErr model.Error
		if errors.As(err, &testErr) {
			var gotRPCModelErr model.Error
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
