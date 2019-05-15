// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcjson_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/daglabs/btcd/util/subnetworkid"

	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/util/daghash"
)

// TestDAGSvrWsNtfns tests all of the dag server websocket-specific
// notifications marshal and unmarshal into valid results include handling of
// optional fields being omitted in the marshalled command, while optional
// fields with defaults have the default assigned on unmarshalled commands.
func TestDAGSvrWsNtfns(t *testing.T) {
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
				return btcjson.NewCmd("filteredBlockAdded", 100000, "header", []string{"tx0", "tx1"})
			},
			staticNtfn: func() interface{} {
				return btcjson.NewFilteredBlockAddedNtfn(100000, "header", []string{"tx0", "tx1"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"filteredBlockAdded","params":[100000,"header",["tx0","tx1"]],"id":null}`,
			unmarshalled: &btcjson.FilteredBlockAddedNtfn{
				Height:        100000,
				Header:        "header",
				SubscribedTxs: []string{"tx0", "tx1"},
			},
		},
		{
			name: "txAccepted",
			newNtfn: func() (interface{}, error) {
				return btcjson.NewCmd("txAccepted", "123", 1.5)
			},
			staticNtfn: func() interface{} {
				return btcjson.NewTxAcceptedNtfn("123", 1.5)
			},
			marshalled: `{"jsonrpc":"1.0","method":"txAccepted","params":["123",1.5],"id":null}`,
			unmarshalled: &btcjson.TxAcceptedNtfn{
				TxID:   "123",
				Amount: 1.5,
			},
		},
		{
			name: "txAcceptedVerbose",
			newNtfn: func() (interface{}, error) {
				return btcjson.NewCmd("txAcceptedVerbose", `{"hex":"001122","txid":"123","version":1,"locktime":4294967295,"subnetwork":"0000000000000000000000000000000000000000","gas":0,"payloadHash":"","payload":"","vin":null,"vout":null,"confirmations":0}`)
			},
			staticNtfn: func() interface{} {
				txResult := btcjson.TxRawResult{
					Hex:           "001122",
					TxID:          "123",
					Version:       1,
					LockTime:      4294967295,
					Subnetwork:    subnetworkid.SubnetworkIDNative.String(),
					Vin:           nil,
					Vout:          nil,
					Confirmations: 0,
				}
				return btcjson.NewTxAcceptedVerboseNtfn(txResult)
			},
			marshalled: `{"jsonrpc":"1.0","method":"txAcceptedVerbose","params":[{"hex":"001122","txId":"123","version":1,"lockTime":4294967295,"subnetwork":"0000000000000000000000000000000000000000","gas":0,"payloadHash":"","payload":"","vin":null,"vout":null,"acceptedBy":null}],"id":null}`,
			unmarshalled: &btcjson.TxAcceptedVerboseNtfn{
				RawTx: btcjson.TxRawResult{
					Hex:           "001122",
					TxID:          "123",
					Version:       1,
					LockTime:      4294967295,
					Subnetwork:    subnetworkid.SubnetworkIDNative.String(),
					Vin:           nil,
					Vout:          nil,
					Confirmations: 0,
				},
			},
		},
		{
			name: "txAcceptedVerbose with subnetwork, gas and paylaod",
			newNtfn: func() (interface{}, error) {
				return btcjson.NewCmd("txAcceptedVerbose", `{"hex":"001122","txId":"123","version":1,"lockTime":4294967295,"subnetwork":"000000000000000000000000000000000000432d","gas":10,"payloadHash":"bf8ccdb364499a3e628200c3d3512c2c2a43b7a7d4f1a40d7f716715e449f442","payload":"102030","vin":null,"vout":null,"acceptedBy":null}`)
			},
			staticNtfn: func() interface{} {
				txResult := btcjson.TxRawResult{
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
					Confirmations: 0,
				}
				return btcjson.NewTxAcceptedVerboseNtfn(txResult)
			},
			marshalled: `{"jsonrpc":"1.0","method":"txAcceptedVerbose","params":[{"hex":"001122","txId":"123","version":1,"lockTime":4294967295,"subnetwork":"000000000000000000000000000000000000432d","gas":10,"payloadHash":"bf8ccdb364499a3e628200c3d3512c2c2a43b7a7d4f1a40d7f716715e449f442","payload":"102030","vin":null,"vout":null,"acceptedBy":null}],"id":null}`,
			unmarshalled: &btcjson.TxAcceptedVerboseNtfn{
				RawTx: btcjson.TxRawResult{
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
					Confirmations: 0,
				},
			},
		},
		{
			name: "relevantTxAccepted",
			newNtfn: func() (interface{}, error) {
				return btcjson.NewCmd("relevantTxAccepted", "001122")
			},
			staticNtfn: func() interface{} {
				return btcjson.NewRelevantTxAcceptedNtfn("001122")
			},
			marshalled: `{"jsonrpc":"1.0","method":"relevantTxAccepted","params":["001122"],"id":null}`,
			unmarshalled: &btcjson.RelevantTxAcceptedNtfn{
				Transaction: "001122",
			},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Marshal the notification as created by the new static
		// creation function.  The ID is nil for notifications.
		marshalled, err := btcjson.MarshalCmd(nil, test.staticNtfn())
		if err != nil {
			t.Errorf("MarshalCmd #%d (%s) unexpected error: %v", i,
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
			t.Errorf("Test #%d (%s) unexpected NewCmd error: %v ",
				i, test.name, err)
		}

		// Marshal the notification as created by the generic new
		// notification creation function.    The ID is nil for
		// notifications.
		marshalled, err = btcjson.MarshalCmd(nil, cmd)
		if err != nil {
			t.Errorf("MarshalCmd #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !bytes.Equal(marshalled, []byte(test.marshalled)) {
			t.Errorf("Test #%d (%s) unexpected marshalled data - "+
				"got %s, want %s", i, test.name, marshalled,
				test.marshalled)
			continue
		}

		var request btcjson.Request
		if err := json.Unmarshal(marshalled, &request); err != nil {
			t.Errorf("Test #%d (%s) unexpected error while "+
				"unmarshalling JSON-RPC request: %v", i,
				test.name, err)
			continue
		}

		cmd, err = btcjson.UnmarshalCmd(&request)
		if err != nil {
			t.Errorf("UnmarshalCmd #%d (%s) unexpected error: %v", i,
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
