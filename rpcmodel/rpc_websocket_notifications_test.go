// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcmodel_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/util/subnetworkid"

	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/daghash"
)

// TestRPCServerWebsocketNotifications tests all of the kaspa rpc server websocket-specific
// notifications marshal and unmarshal into valid results include handling of
// optional fields being omitted in the marshalled command, while optional
// fields with defaults have the default assigned on unmarshalled commands.
func TestRPCServerWebsocketNotifications(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		newNtfn      func() (interface{}, error)
		staticNtfn   func() interface{}
		marshalled   string
		unmarshalled interface{}
	}{
		{
			name: "filteredBlockAdded",
			newNtfn: func() (interface{}, error) {
				return rpcmodel.NewCommand("filteredBlockAdded", 100, "header", []string{"tx0", "tx1"})
			},
			staticNtfn: func() interface{} {
				return rpcmodel.NewFilteredBlockAddedNtfn(100, "header", []string{"tx0", "tx1"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"filteredBlockAdded","params":[100,"header",["tx0","tx1"]],"id":null}`,
			unmarshalled: &rpcmodel.FilteredBlockAddedNtfn{
				BlueScore:     100,
				Header:        "header",
				SubscribedTxs: []string{"tx0", "tx1"},
			},
		},
		{
			name: "txAccepted",
			newNtfn: func() (interface{}, error) {
				return rpcmodel.NewCommand("txAccepted", "123", 1.5)
			},
			staticNtfn: func() interface{} {
				return rpcmodel.NewTxAcceptedNtfn("123", 1.5)
			},
			marshalled: `{"jsonrpc":"1.0","method":"txAccepted","params":["123",1.5],"id":null}`,
			unmarshalled: &rpcmodel.TxAcceptedNtfn{
				TxID:   "123",
				Amount: 1.5,
			},
		},
		{
			name: "txAcceptedVerbose",
			newNtfn: func() (interface{}, error) {
				return rpcmodel.NewCommand("txAcceptedVerbose", `{"hex":"001122","txid":"123","version":1,"locktime":4294967295,"subnetwork":"0000000000000000000000000000000000000000","gas":0,"payloadHash":"","payload":"","vin":null,"vout":null,"isInMempool":false}`)
			},
			staticNtfn: func() interface{} {
				txResult := rpcmodel.TxRawResult{
					Hex:           "001122",
					TxID:          "123",
					Version:       1,
					LockTime:      4294967295,
					Subnetwork:    subnetworkid.SubnetworkIDNative.String(),
					Vin:           nil,
					Vout:          nil,
					Confirmations: nil,
				}
				return rpcmodel.NewTxAcceptedVerboseNtfn(txResult)
			},
			marshalled: `{"jsonrpc":"1.0","method":"txAcceptedVerbose","params":[{"hex":"001122","txId":"123","version":1,"lockTime":4294967295,"subnetwork":"0000000000000000000000000000000000000000","gas":0,"payloadHash":"","payload":"","vin":null,"vout":null,"isInMempool":false}],"id":null}`,
			unmarshalled: &rpcmodel.TxAcceptedVerboseNtfn{
				RawTx: rpcmodel.TxRawResult{
					Hex:           "001122",
					TxID:          "123",
					Version:       1,
					LockTime:      4294967295,
					Subnetwork:    subnetworkid.SubnetworkIDNative.String(),
					Vin:           nil,
					Vout:          nil,
					Confirmations: nil,
				},
			},
		},
		{
			name: "txAcceptedVerbose with subnetwork, gas and paylaod",
			newNtfn: func() (interface{}, error) {
				return rpcmodel.NewCommand("txAcceptedVerbose", `{"hex":"001122","txId":"123","version":1,"lockTime":4294967295,"subnetwork":"000000000000000000000000000000000000432d","gas":10,"payloadHash":"bf8ccdb364499a3e628200c3d3512c2c2a43b7a7d4f1a40d7f716715e449f442","payload":"102030","vin":null,"vout":null,"isInMempool":false}`)
			},
			staticNtfn: func() interface{} {
				txResult := rpcmodel.TxRawResult{
					Hex:           "001122",
					TxID:          "123",
					Version:       1,
					LockTime:      4294967295,
					Subnetwork:    subnetworkid.SubnetworkID{45, 67}.String(),
					PayloadHash:   daghash.DoubleHashP([]byte("102030")).String(),
					Payload:       "102030",
					Gas:           10,
					Vin:           nil,
					Vout:          nil,
					Confirmations: nil,
				}
				return rpcmodel.NewTxAcceptedVerboseNtfn(txResult)
			},
			marshalled: `{"jsonrpc":"1.0","method":"txAcceptedVerbose","params":[{"hex":"001122","txId":"123","version":1,"lockTime":4294967295,"subnetwork":"000000000000000000000000000000000000432d","gas":10,"payloadHash":"bf8ccdb364499a3e628200c3d3512c2c2a43b7a7d4f1a40d7f716715e449f442","payload":"102030","vin":null,"vout":null,"isInMempool":false}],"id":null}`,
			unmarshalled: &rpcmodel.TxAcceptedVerboseNtfn{
				RawTx: rpcmodel.TxRawResult{
					Hex:           "001122",
					TxID:          "123",
					Version:       1,
					LockTime:      4294967295,
					Subnetwork:    subnetworkid.SubnetworkID{45, 67}.String(),
					PayloadHash:   daghash.DoubleHashP([]byte("102030")).String(),
					Payload:       "102030",
					Gas:           10,
					Vin:           nil,
					Vout:          nil,
					Confirmations: nil,
				},
			},
		},
		{
			name: "relevantTxAccepted",
			newNtfn: func() (interface{}, error) {
				return rpcmodel.NewCommand("relevantTxAccepted", "001122")
			},
			staticNtfn: func() interface{} {
				return rpcmodel.NewRelevantTxAcceptedNtfn("001122")
			},
			marshalled: `{"jsonrpc":"1.0","method":"relevantTxAccepted","params":["001122"],"id":null}`,
			unmarshalled: &rpcmodel.RelevantTxAcceptedNtfn{
				Transaction: "001122",
			},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Marshal the notification as created by the new static
		// creation function. The ID is nil for notifications.
		marshalled, err := rpcmodel.MarshalCommand(nil, test.staticNtfn())
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

		// Ensure the notification is created without error via the
		// generic new notification creation function.
		cmd, err := test.newNtfn()
		if err != nil {
			t.Errorf("Test #%d (%s) unexpected NewCommand error: %v ",
				i, test.name, err)
		}

		// Marshal the notification as created by the generic new
		// notification creation function. The ID is nil for
		// notifications.
		marshalled, err = rpcmodel.MarshalCommand(nil, cmd)
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
