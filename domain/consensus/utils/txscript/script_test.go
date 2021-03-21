// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"bytes"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"reflect"
	"testing"
)

// TestParseOpcode tests for opcode parsing with bad data templates.
func TestParseOpcode(t *testing.T) {
	// Deep copy the array and make one of the opcodes invalid by setting it
	// to the wrong length.
	fakeArray := opcodeArray
	fakeArray[OpPushData4] = opcode{value: OpPushData4,
		name: "OP_PUSHDATA4", length: -8, opfunc: opcodePushData}

	// This script would be fine if -8 was a valid length.
	_, err := parseScriptTemplate([]byte{OpPushData4, 0x1, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00}, &fakeArray)
	if err == nil {
		t.Errorf("no error with dodgy opcode array!")
	}
}

// TestUnparsingInvalidOpcodes tests for errors when unparsing invalid parsed
// opcodes.
func TestUnparsingInvalidOpcodes(t *testing.T) {
	tests := []struct {
		name        string
		pop         *parsedOpcode
		expectedErr error
	}{
		{
			name: "OP_FALSE",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpFalse],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_FALSE long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpFalse],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_1 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData1],
				data:   nil,
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_1",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData1],
				data:   make([]byte, 1),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_1 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData1],
				data:   make([]byte, 2),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_2 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData2],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_2",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData2],
				data:   make([]byte, 2),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_2 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData2],
				data:   make([]byte, 3),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_3 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData3],
				data:   make([]byte, 2),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_3",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData3],
				data:   make([]byte, 3),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_3 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData3],
				data:   make([]byte, 4),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_4 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData4],
				data:   make([]byte, 3),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_4",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData4],
				data:   make([]byte, 4),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_4 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData4],
				data:   make([]byte, 5),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_5 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData5],
				data:   make([]byte, 4),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_5",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData5],
				data:   make([]byte, 5),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_5 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData5],
				data:   make([]byte, 6),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_6 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData6],
				data:   make([]byte, 5),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_6",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData6],
				data:   make([]byte, 6),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_6 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData6],
				data:   make([]byte, 7),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_7 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData7],
				data:   make([]byte, 6),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_7",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData7],
				data:   make([]byte, 7),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_7 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData7],
				data:   make([]byte, 8),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_8 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData8],
				data:   make([]byte, 7),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_8",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData8],
				data:   make([]byte, 8),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_8 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData8],
				data:   make([]byte, 9),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_9 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData9],
				data:   make([]byte, 8),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_9",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData9],
				data:   make([]byte, 9),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_9 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData9],
				data:   make([]byte, 10),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_10 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData10],
				data:   make([]byte, 9),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_10",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData10],
				data:   make([]byte, 10),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_10 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData10],
				data:   make([]byte, 11),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_11 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData11],
				data:   make([]byte, 10),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_11",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData11],
				data:   make([]byte, 11),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_11 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData11],
				data:   make([]byte, 12),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_12 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData12],
				data:   make([]byte, 11),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_12",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData12],
				data:   make([]byte, 12),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_12 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData12],
				data:   make([]byte, 13),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_13 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData13],
				data:   make([]byte, 12),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_13",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData13],
				data:   make([]byte, 13),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_13 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData13],
				data:   make([]byte, 14),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_14 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData14],
				data:   make([]byte, 13),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_14",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData14],
				data:   make([]byte, 14),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_14 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData14],
				data:   make([]byte, 15),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_15 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData15],
				data:   make([]byte, 14),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_15",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData15],
				data:   make([]byte, 15),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_15 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData15],
				data:   make([]byte, 16),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_16 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData16],
				data:   make([]byte, 15),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_16",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData16],
				data:   make([]byte, 16),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_16 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData16],
				data:   make([]byte, 17),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_17 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData17],
				data:   make([]byte, 16),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_17",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData17],
				data:   make([]byte, 17),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_17 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData17],
				data:   make([]byte, 18),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_18 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData18],
				data:   make([]byte, 17),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_18",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData18],
				data:   make([]byte, 18),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_18 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData18],
				data:   make([]byte, 19),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_19 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData19],
				data:   make([]byte, 18),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_19",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData19],
				data:   make([]byte, 19),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_19 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData19],
				data:   make([]byte, 20),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_20 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData20],
				data:   make([]byte, 19),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_20",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData20],
				data:   make([]byte, 20),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_20 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData20],
				data:   make([]byte, 21),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_21 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData21],
				data:   make([]byte, 20),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_21",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData21],
				data:   make([]byte, 21),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_21 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData21],
				data:   make([]byte, 22),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_22 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData22],
				data:   make([]byte, 21),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_22",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData22],
				data:   make([]byte, 22),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_22 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData22],
				data:   make([]byte, 23),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_23 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData23],
				data:   make([]byte, 22),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_23",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData23],
				data:   make([]byte, 23),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_23 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData23],
				data:   make([]byte, 24),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_24 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData24],
				data:   make([]byte, 23),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_24",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData24],
				data:   make([]byte, 24),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_24 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData24],
				data:   make([]byte, 25),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_25 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData25],
				data:   make([]byte, 24),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_25",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData25],
				data:   make([]byte, 25),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_25 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData25],
				data:   make([]byte, 26),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_26 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData26],
				data:   make([]byte, 25),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_26",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData26],
				data:   make([]byte, 26),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_26 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData26],
				data:   make([]byte, 27),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_27 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData27],
				data:   make([]byte, 26),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_27",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData27],
				data:   make([]byte, 27),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_27 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData27],
				data:   make([]byte, 28),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_28 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData28],
				data:   make([]byte, 27),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_28",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData28],
				data:   make([]byte, 28),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_28 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData28],
				data:   make([]byte, 29),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_29 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData29],
				data:   make([]byte, 28),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_29",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData29],
				data:   make([]byte, 29),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_29 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData29],
				data:   make([]byte, 30),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_30 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData30],
				data:   make([]byte, 29),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_30",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData30],
				data:   make([]byte, 30),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_30 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData30],
				data:   make([]byte, 31),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_31 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData31],
				data:   make([]byte, 30),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_31",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData31],
				data:   make([]byte, 31),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_31 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData31],
				data:   make([]byte, 32),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_32 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData32],
				data:   make([]byte, 31),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_32",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData32],
				data:   make([]byte, 32),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_32 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData32],
				data:   make([]byte, 33),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_33 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData33],
				data:   make([]byte, 32),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_33",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData33],
				data:   make([]byte, 33),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_33 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData33],
				data:   make([]byte, 34),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_34 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData34],
				data:   make([]byte, 33),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_34",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData34],
				data:   make([]byte, 34),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_34 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData34],
				data:   make([]byte, 35),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_35 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData35],
				data:   make([]byte, 34),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_35",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData35],
				data:   make([]byte, 35),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_35 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData35],
				data:   make([]byte, 36),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_36 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData36],
				data:   make([]byte, 35),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_36",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData36],
				data:   make([]byte, 36),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_36 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData36],
				data:   make([]byte, 37),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_37 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData37],
				data:   make([]byte, 36),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_37",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData37],
				data:   make([]byte, 37),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_37 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData37],
				data:   make([]byte, 38),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_38 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData38],
				data:   make([]byte, 37),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_38",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData38],
				data:   make([]byte, 38),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_38 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData38],
				data:   make([]byte, 39),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_39 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData39],
				data:   make([]byte, 38),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_39",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData39],
				data:   make([]byte, 39),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_39 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData39],
				data:   make([]byte, 40),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_40 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData40],
				data:   make([]byte, 39),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_40",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData40],
				data:   make([]byte, 40),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_40 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData40],
				data:   make([]byte, 41),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_41 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData41],
				data:   make([]byte, 40),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_41",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData41],
				data:   make([]byte, 41),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_41 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData41],
				data:   make([]byte, 42),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_42 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData42],
				data:   make([]byte, 41),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_42",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData42],
				data:   make([]byte, 42),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_42 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData42],
				data:   make([]byte, 43),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_43 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData43],
				data:   make([]byte, 42),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_43",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData43],
				data:   make([]byte, 43),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_43 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData43],
				data:   make([]byte, 44),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_44 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData44],
				data:   make([]byte, 43),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_44",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData44],
				data:   make([]byte, 44),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_44 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData44],
				data:   make([]byte, 45),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_45 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData45],
				data:   make([]byte, 44),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_45",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData45],
				data:   make([]byte, 45),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_45 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData45],
				data:   make([]byte, 46),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_46 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData46],
				data:   make([]byte, 45),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_46",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData46],
				data:   make([]byte, 46),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_46 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData46],
				data:   make([]byte, 47),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_47 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData47],
				data:   make([]byte, 46),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_47",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData47],
				data:   make([]byte, 47),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_47 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData47],
				data:   make([]byte, 48),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_48 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData48],
				data:   make([]byte, 47),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_48",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData48],
				data:   make([]byte, 48),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_48 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData48],
				data:   make([]byte, 49),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_49 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData49],
				data:   make([]byte, 48),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_49",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData49],
				data:   make([]byte, 49),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_49 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData49],
				data:   make([]byte, 50),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_50 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData50],
				data:   make([]byte, 49),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_50",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData50],
				data:   make([]byte, 50),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_50 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData50],
				data:   make([]byte, 51),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_51 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData51],
				data:   make([]byte, 50),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_51",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData51],
				data:   make([]byte, 51),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_51 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData51],
				data:   make([]byte, 52),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_52 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData52],
				data:   make([]byte, 51),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_52",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData52],
				data:   make([]byte, 52),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_52 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData52],
				data:   make([]byte, 53),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_53 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData53],
				data:   make([]byte, 52),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_53",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData53],
				data:   make([]byte, 53),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_53 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData53],
				data:   make([]byte, 54),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_54 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData54],
				data:   make([]byte, 53),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_54",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData54],
				data:   make([]byte, 54),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_54 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData54],
				data:   make([]byte, 55),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_55 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData55],
				data:   make([]byte, 54),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_55",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData55],
				data:   make([]byte, 55),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_55 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData55],
				data:   make([]byte, 56),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_56 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData56],
				data:   make([]byte, 55),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_56",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData56],
				data:   make([]byte, 56),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_56 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData56],
				data:   make([]byte, 57),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_57 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData57],
				data:   make([]byte, 56),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_57",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData57],
				data:   make([]byte, 57),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_57 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData57],
				data:   make([]byte, 58),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_58 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData58],
				data:   make([]byte, 57),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_58",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData58],
				data:   make([]byte, 58),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_58 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData58],
				data:   make([]byte, 59),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_59 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData59],
				data:   make([]byte, 58),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_59",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData59],
				data:   make([]byte, 59),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_59 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData59],
				data:   make([]byte, 60),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_60 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData60],
				data:   make([]byte, 59),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_60",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData60],
				data:   make([]byte, 60),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_60 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData60],
				data:   make([]byte, 61),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_61 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData61],
				data:   make([]byte, 60),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_61",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData61],
				data:   make([]byte, 61),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_61 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData61],
				data:   make([]byte, 62),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_62 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData62],
				data:   make([]byte, 61),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_62",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData62],
				data:   make([]byte, 62),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_62 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData62],
				data:   make([]byte, 63),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_63 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData63],
				data:   make([]byte, 62),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_63",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData63],
				data:   make([]byte, 63),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_63 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData63],
				data:   make([]byte, 64),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_64 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData64],
				data:   make([]byte, 63),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_64",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData64],
				data:   make([]byte, 64),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_64 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData64],
				data:   make([]byte, 65),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_65 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData65],
				data:   make([]byte, 64),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_65",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData65],
				data:   make([]byte, 65),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_65 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData65],
				data:   make([]byte, 66),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_66 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData66],
				data:   make([]byte, 65),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_66",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData66],
				data:   make([]byte, 66),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_66 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData66],
				data:   make([]byte, 67),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_67 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData67],
				data:   make([]byte, 66),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_67",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData67],
				data:   make([]byte, 67),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_67 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData67],
				data:   make([]byte, 68),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_68 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData68],
				data:   make([]byte, 67),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_68",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData68],
				data:   make([]byte, 68),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_68 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData68],
				data:   make([]byte, 69),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_69 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData69],
				data:   make([]byte, 68),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_69",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData69],
				data:   make([]byte, 69),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_69 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData69],
				data:   make([]byte, 70),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_70 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData70],
				data:   make([]byte, 69),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_70",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData70],
				data:   make([]byte, 70),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_70 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData70],
				data:   make([]byte, 71),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_71 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData71],
				data:   make([]byte, 70),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_71",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData71],
				data:   make([]byte, 71),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_71 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData71],
				data:   make([]byte, 72),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_72 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData72],
				data:   make([]byte, 71),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_72",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData72],
				data:   make([]byte, 72),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_72 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData72],
				data:   make([]byte, 73),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_73 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData73],
				data:   make([]byte, 72),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_73",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData73],
				data:   make([]byte, 73),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_73 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData73],
				data:   make([]byte, 74),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_74 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData74],
				data:   make([]byte, 73),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_74",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData74],
				data:   make([]byte, 74),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_74 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData74],
				data:   make([]byte, 75),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_75 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData75],
				data:   make([]byte, 74),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DATA_75",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData75],
				data:   make([]byte, 75),
			},
			expectedErr: nil,
		},
		{
			name: "OP_DATA_75 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData75],
				data:   make([]byte, 76),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_PUSHDATA1",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPushData1],
				data:   []byte{0, 1, 2, 3, 4},
			},
			expectedErr: nil,
		},
		{
			name: "OP_PUSHDATA2",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPushData2],
				data:   []byte{0, 1, 2, 3, 4},
			},
			expectedErr: nil,
		},
		{
			name: "OP_PUSHDATA4",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPushData1],
				data:   []byte{0, 1, 2, 3, 4},
			},
			expectedErr: nil,
		},
		{
			name: "OP_1NEGATE",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Negate],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_1NEGATE long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Negate],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_RESERVED",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReserved],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_RESERVED long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReserved],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_TRUE",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpTrue],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_TRUE long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpTrue],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_2",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_2 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_2",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_2 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_3",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op3],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_3 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op3],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_4",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op4],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_4 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op4],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_5",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op5],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_5 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op5],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_6",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op6],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_6 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op6],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_7",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op7],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_7 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op7],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_8",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op8],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_8 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op8],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_9",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op9],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_9 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op9],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_10",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op10],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_10 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op10],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_11",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op11],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_11 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op11],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_12",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op12],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_12 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op12],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_13",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op13],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_13 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op13],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_14",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op14],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_14 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op14],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_15",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op15],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_15 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op15],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_16",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op16],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_16 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op16],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NOP",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NOP long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_VER",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVer],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_VER long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVer],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_IF",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpIf],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_IF long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpIf],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NOTIF",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNotIf],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NOTIF long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNotIf],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_VERIF",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVerIf],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_VERIF long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVerIf],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_VERNOTIF",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVerNotIf],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_VERNOTIF long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVerNotIf],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_ELSE",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpElse],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_ELSE long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpElse],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_ENDIF",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpEndIf],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_ENDIF long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpEndIf],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_VERIFY",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVerify],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_VERIFY long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVerify],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_RETURN",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReturn],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_RETURN long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReturn],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_TOALTSTACK",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpToAltStack],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_TOALTSTACK long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpToAltStack],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_FROMALTSTACK",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpFromAltStack],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_FROMALTSTACK long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpFromAltStack],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_2DROP",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Drop],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_2DROP long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Drop],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_2DUP",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Dup],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_2DUP long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Dup],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_3DUP",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op3Dup],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_3DUP long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op3Dup],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_2OVER",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Over],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_2OVER long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Over],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_2ROT",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Rot],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_2ROT long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Rot],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_2SWAP",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Swap],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_2SWAP long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Swap],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_IFDUP",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpIfDup],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_IFDUP long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpIfDup],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DEPTH",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDepth],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_DEPTH long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDepth],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DROP",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDrop],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_DROP long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDrop],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DUP",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDup],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_DUP long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDup],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NIP",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNip],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NIP long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNip],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_OVER",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpOver],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_OVER long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpOver],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_PICK",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPick],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_PICK long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPick],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_ROLL",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRoll],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_ROLL long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRoll],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_ROT",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRot],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_ROT long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRot],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_SWAP",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSwap],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_SWAP long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSwap],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_TUCK",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpTuck],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_TUCK long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpTuck],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_CAT",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCat],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_CAT long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCat],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_SUBSTR",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSubStr],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_SUBSTR long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSubStr],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_LEFT",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLeft],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_LEFT long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLeft],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_LEFT",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLeft],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_LEFT long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLeft],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_RIGHT",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRight],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_RIGHT long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRight],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_SIZE",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSize],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_SIZE long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSize],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_INVERT",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpInvert],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_INVERT long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpInvert],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_AND",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpAnd],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_AND long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpAnd],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_OR",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpOr],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_OR long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpOr],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_XOR",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpXor],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_XOR long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpXor],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_EQUAL",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpEqual],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_EQUAL long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpEqual],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_EQUALVERIFY",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpEqualVerify],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_EQUALVERIFY long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpEqualVerify],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_RESERVED1",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReserved1],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_RESERVED1 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReserved1],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_RESERVED2",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReserved2],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_RESERVED2 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReserved2],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_1ADD",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Add],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_1ADD long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Add],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_1SUB",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Sub],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_1SUB long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Sub],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_2MUL",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Mul],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_2MUL long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Mul],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_2DIV",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Div],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_2DIV long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Div],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NEGATE",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNegate],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NEGATE long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNegate],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_ABS",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpAbs],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_ABS long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpAbs],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NOT",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNot],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NOT long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNot],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_0NOTEQUAL",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op0NotEqual],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_0NOTEQUAL long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op0NotEqual],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_ADD",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpAdd],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_ADD long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpAdd],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_SUB",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSub],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_SUB long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSub],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_MUL",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMul],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_MUL long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMul],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_DIV",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDiv],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_DIV long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDiv],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_MOD",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMod],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_MOD long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMod],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_LSHIFT",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLShift],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_LSHIFT long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLShift],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_RSHIFT",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRShift],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_RSHIFT long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRShift],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_BOOLAND",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpBoolAnd],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_BOOLAND long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpBoolAnd],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_BOOLOR",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpBoolOr],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_BOOLOR long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpBoolOr],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NUMEQUAL",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNumEqual],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NUMEQUAL long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNumEqual],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NUMEQUALVERIFY",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNumEqualVerify],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NUMEQUALVERIFY long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNumEqualVerify],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NUMNOTEQUAL",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNumNotEqual],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NUMNOTEQUAL long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNumNotEqual],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_LESSTHAN",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLessThan],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_LESSTHAN long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLessThan],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_GREATERTHAN",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpGreaterThan],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_GREATERTHAN long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpGreaterThan],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_LESSTHANOREQUAL",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLessThanOrEqual],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_LESSTHANOREQUAL long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLessThanOrEqual],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_GREATERTHANOREQUAL",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpGreaterThanOrEqual],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_GREATERTHANOREQUAL long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpGreaterThanOrEqual],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_MIN",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMin],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_MIN long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMin],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_MAX",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMax],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_MAX long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMax],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_WITHIN",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpWithin],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_WITHIN long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpWithin],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_SHA256",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSHA256],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_SHA256 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSHA256],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_BLAKE2B",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpBlake2b],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_BLAKE2B long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpBlake2b],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_CHECKSIG",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckSig],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_CHECKSIG long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckSig],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_CHECKSIGVERIFY",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckSigVerify],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_CHECKSIGVERIFY long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckSigVerify],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_CHECKMULTISIG",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckMultiSig],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_CHECKMULTISIG long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckMultiSig],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_CHECKMULTISIGVERIFY",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckMultiSigVerify],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_CHECKMULTISIGVERIFY long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckMultiSigVerify],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NOP1",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop1],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NOP1 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop1],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NOP2",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop2],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NOP2 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop2],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NOP3",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop3],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NOP3 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop3],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NOP4",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop4],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NOP4 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop4],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NOP5",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop5],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NOP5 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop5],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NOP6",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop6],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NOP6 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop6],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NOP7",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop7],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NOP7 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop7],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NOP8",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop8],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NOP8 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop8],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NOP9",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop9],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NOP9 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop9],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_NOP10",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop10],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_NOP10 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNop10],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_PUBKEYHASH",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPubKeyHash],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_PUBKEYHASH long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPubKeyHash],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_PUBKEY",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPubKey],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_PUBKEY long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPubKey],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_INVALIDOPCODE",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpInvalidOpCode],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OP_INVALIDOPCODE long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpInvalidOpCode],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
	}

	for _, test := range tests {
		_, err := test.pop.bytes()
		if e := checkScriptError(err, test.expectedErr); e != nil {
			t.Errorf("Parsed opcode test '%s': %v", test.name, e)
			continue
		}
	}
}

// TestPushedData ensured the PushedData function extracts the expected data out
// of various scripts.
func TestPushedData(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		script string
		out    [][]byte
		valid  bool
	}{
		{
			"0 IF 0 ELSE 2 ENDIF",
			[][]byte{nil, nil},
			true,
		},
		{
			"16777216 10000000",
			[][]byte{
				{0x00, 0x00, 0x00, 0x01}, // 16777216
				{0x80, 0x96, 0x98, 0x00}, // 10000000
			},
			true,
		},
		{
			"DUP BLAKE2B '17VZNX1SN5NtKa8UQFxwQbFeFc3iqRYhem' EQUALVERIFY CHECKSIG",
			[][]byte{
				// 17VZNX1SN5NtKa8UQFxwQbFeFc3iqRYhem
				{
					0x31, 0x37, 0x56, 0x5a, 0x4e, 0x58, 0x31, 0x53, 0x4e, 0x35,
					0x4e, 0x74, 0x4b, 0x61, 0x38, 0x55, 0x51, 0x46, 0x78, 0x77,
					0x51, 0x62, 0x46, 0x65, 0x46, 0x63, 0x33, 0x69, 0x71, 0x52,
					0x59, 0x68, 0x65, 0x6d,
				},
			},
			true,
		},
		{
			"PUSHDATA4 1000 EQUAL",
			nil,
			false,
		},
	}

	for i, test := range tests {
		script := mustParseShortForm(test.script, 0)
		data, err := PushedData(script)
		if test.valid && err != nil {
			t.Errorf("TestPushedData failed test #%d: %v\n", i, err)
			continue
		} else if !test.valid && err == nil {
			t.Errorf("TestPushedData failed test #%d: test should "+
				"be invalid\n", i)
			continue
		}
		if !reflect.DeepEqual(data, test.out) {
			t.Errorf("TestPushedData failed test #%d: want: %x "+
				"got: %x\n", i, test.out, data)
		}
	}
}

// isPushOnlyScript returns whether or not the passed script only pushes data.
//
// False will be returned when the script does not parse.
func isPushOnlyScript(script []byte) (bool, error) {
	pops, err := parseScript(script)
	if err != nil {
		return false, err
	}
	return isPushOnly(pops), nil
}

// TestHasCanonicalPush ensures the canonicalPush function works as expected.
func TestHasCanonicalPush(t *testing.T) {
	t.Parallel()

	for i := 0; i < 65535; i++ {
		script, err := NewScriptBuilder().AddInt64(int64(i)).Script()
		if err != nil {
			t.Errorf("Script: test #%d unexpected error: %v\n", i,
				err)
			continue
		}
		if result, _ := isPushOnlyScript(script); !result {
			t.Errorf("isPushOnlyScript: test #%d failed: %x\n", i,
				script)
			continue
		}
		pops, err := parseScript(script)
		if err != nil {
			t.Errorf("parseScript: #%d failed: %v", i, err)
			continue
		}
		for _, pop := range pops {
			if result := canonicalPush(pop); !result {
				t.Errorf("canonicalPush: test #%d failed: %x\n",
					i, script)
				break
			}
		}
	}
	for i := 0; i <= MaxScriptElementSize; i++ {
		builder := NewScriptBuilder()
		builder.AddData(bytes.Repeat([]byte{0x49}, i))
		script, err := builder.Script()
		if err != nil {
			t.Errorf("StandardPushesTests test #%d unexpected error: %v\n", i, err)
			continue
		}
		if result, _ := isPushOnlyScript(script); !result {
			t.Errorf("StandardPushesTests isPushOnlyScript test #%d failed: %x\n", i, script)
			continue
		}
		pops, err := parseScript(script)
		if err != nil {
			t.Errorf("StandardPushesTests #%d failed to TstParseScript: %v", i, err)
			continue
		}
		for _, pop := range pops {
			if result := canonicalPush(pop); !result {
				t.Errorf("StandardPushesTests TstHasCanonicalPushes test #%d failed: %x\n", i, script)
				break
			}
		}
	}
}

// TestGetPreciseSigOps ensures the more precise signature operation counting
// mechanism which includes signatures in P2SH scripts works as expected.
func TestGetPreciseSigOps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		scriptSig []byte
		nSigOps   int
	}{
		{
			name:      "scriptSig doesn't parse",
			scriptSig: mustParseShortForm("PUSHDATA1 0x02", 0),
		},
		{
			name:      "scriptSig isn't push only",
			scriptSig: mustParseShortForm("1 DUP", 0),
			nSigOps:   0,
		},
		{
			name:      "scriptSig length 0",
			scriptSig: nil,
			nSigOps:   0,
		},
		{
			name: "No script at the end",
			// No script at end but still push only.
			scriptSig: mustParseShortForm("1 1", 0),
			nSigOps:   0,
		},
		{
			name:      "pushed script doesn't parse",
			scriptSig: mustParseShortForm("DATA_2 PUSHDATA1 0x02", 0),
		},
	}

	// The signature in the p2sh script is nonsensical for the tests since
	// this script will never be executed. What matters is that it matches
	// the right pattern.
	scriptOnly := mustParseShortForm("BLAKE2B DATA_32 0x433ec2ac1ffa1b7b7d0"+
		"27f564529c57197f9ae88 EQUAL", 0)
	scriptPubKey := &externalapi.ScriptPublicKey{scriptOnly, 0}
	for _, test := range tests {
		count := GetPreciseSigOpCount(test.scriptSig, scriptPubKey, true)
		if count != test.nSigOps {
			t.Errorf("%s: expected count of %d, got %d", test.name,
				test.nSigOps, count)

		}
	}
}

// TestIsPayToScriptHash ensures the IsPayToScriptHash function returns the
// expected results for all the scripts in scriptClassTests.
func TestIsPayToScriptHash(t *testing.T) {
	t.Parallel()

	for _, test := range scriptClassTests {
		script := &externalapi.ScriptPublicKey{mustParseShortForm(test.script, 0), 0}
		shouldBe := (test.class == ScriptHashTy)
		p2sh := IsPayToScriptHash(script)
		if p2sh != shouldBe {
			t.Errorf("%s: expected p2sh %v, got %v", test.name,
				shouldBe, p2sh)
		}
	}
}

// TestHasCanonicalPushes ensures the canonicalPush function properly determines
// what is considered a canonical push.
func TestHasCanonicalPushes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		script   string
		expected bool
	}{
		{
			name: "does not parse",
			script: "0x046708afdb0fe5548271967f1a67130b7105cd6a82" +
				"8e03909a67962e0ea1f61d",
			expected: false,
		},
		{
			name:     "non-canonical push",
			script:   "PUSHDATA1 0x04 0x01020304",
			expected: false,
		},
	}

	for i, test := range tests {
		script := mustParseShortForm(test.script, 0)
		pops, err := parseScript(script)
		if err != nil {
			if test.expected {
				t.Errorf("TstParseScript #%d failed: %v", i, err)
			}
			continue
		}
		for _, pop := range pops {
			if canonicalPush(pop) != test.expected {
				t.Errorf("canonicalPush: #%d (%s) wrong result"+
					"\ngot: %v\nwant: %v", i, test.name,
					true, test.expected)
				break
			}
		}
	}
}

// TestIsPushOnly ensures the isPushOnly function returns the
// expected results.
func TestIsPushOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		script         []byte
		expectedResult bool
		shouldFail     bool
	}{
		{
			name: "does not parse",
			script: mustParseShortForm("0x046708afdb0fe5548271967f1a67130"+
				"b7105cd6a828e03909a67962e0ea1f61d", 0),
			expectedResult: false,
			shouldFail:     true,
		},
		{
			name:           "non push only script",
			script:         mustParseShortForm("0x515293", 0), //OP_1 OP_2 OP_ADD
			expectedResult: false,
			shouldFail:     false,
		},
		{
			name:           "push only script",
			script:         mustParseShortForm("0x5152", 0), //OP_1 OP_2
			expectedResult: true,
			shouldFail:     false,
		},
	}

	for _, test := range tests {
		isPushOnly, err := isPushOnlyScript(test.script)

		if isPushOnly != test.expectedResult {
			t.Errorf("isPushOnlyScript (%s) wrong result\ngot: %v\nwant: "+
				"%v", test.name, isPushOnly, test.expectedResult)
		}

		if test.shouldFail && err == nil {
			t.Errorf("isPushOnlyScript (%s) expected an error but got <nil>", test.name)
		}

		if !test.shouldFail && err != nil {
			t.Errorf("isPushOnlyScript (%s) expected no error but got: %v", test.name, err)
		}
	}
}

// TestIsUnspendable ensures the IsUnspendable function returns the expected
// results.
func TestIsUnspendable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		scriptPubKey []byte
		expected     bool
	}{
		{
			// Unspendable
			scriptPubKey: []byte{0x6a, 0x04, 0x74, 0x65, 0x73, 0x74},
			expected:     true,
		},
		{
			// Spendable
			scriptPubKey: []byte{0x76, 0xa9, 0x14, 0x29, 0x95, 0xa0,
				0xfe, 0x68, 0x43, 0xfa, 0x9b, 0x95, 0x45,
				0x97, 0xf0, 0xdc, 0xa7, 0xa4, 0x4d, 0xf6,
				0xfa, 0x0b, 0x5c, 0x88, 0xac},
			expected: false,
		},
	}

	for i, test := range tests {
		res := IsUnspendable(test.scriptPubKey)
		if res != test.expected {
			t.Errorf("TestIsUnspendable #%d failed: got %v want %v",
				i, res, test.expected)
			continue
		}
	}
}
