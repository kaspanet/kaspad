// Copyright (c) 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package model_test

import (
	"encoding/json"
	"github.com/kaspanet/kaspad/util/pointers"
	"testing"

	"github.com/kaspanet/kaspad/network/rpc/model"
)

// TestRPCServerCustomResults ensures any results that have custom marshalling
// work as intended.
// and unmarshal code of results are as expected.
func TestRPCServerCustomResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   interface{}
		expected string
	}{
		{
			name: "custom vin marshal without coinbase",
			result: &model.Vin{
				TxID: "123",
				Vout: 1,
				ScriptSig: &model.ScriptSig{
					Asm: "0",
					Hex: "00",
				},
				Sequence: 4294967295,
			},
			expected: `{"txId":"123","vout":1,"scriptSig":{"asm":"0","hex":"00"},"sequence":4294967295}`,
		},
		{
			name: "custom vinprevout marshal with coinbase",
			result: &model.VinPrevOut{
				Coinbase: "021234",
				Sequence: 4294967295,
			},
			expected: `{"coinbase":"021234","sequence":4294967295}`,
		},
		{
			name: "custom vinprevout marshal without coinbase",
			result: &model.VinPrevOut{
				TxID: "123",
				Vout: 1,
				ScriptSig: &model.ScriptSig{
					Asm: "0",
					Hex: "00",
				},
				PrevOut: &model.PrevOut{
					Address: pointers.String("addr1"),
					Value:   0,
				},
				Sequence: 4294967295,
			},
			expected: `{"txId":"123","vout":1,"scriptSig":{"asm":"0","hex":"00"},"prevOut":{"address":"addr1","value":0},"sequence":4294967295}`,
		},
		{
			name: "versionresult",
			result: &model.VersionResult{
				VersionString: "1.0.0",
				Major:         1,
				Minor:         0,
				Patch:         0,
				Prerelease:    "pr",
				BuildMetadata: "bm",
			},
			expected: `{"versionString":"1.0.0","major":1,"minor":0,"patch":0,"prerelease":"pr","buildMetadata":"bm"}`,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		marshalled, err := json.Marshal(test.result)
		if err != nil {
			t.Errorf("Test #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}
		if string(marshalled) != test.expected {
			t.Errorf("Test #%d (%s) unexpected marhsalled data - "+
				"got %s, want %s", i, test.name, marshalled,
				test.expected)
			continue
		}
	}
}
