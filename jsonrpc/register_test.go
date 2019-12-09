// Copyright (c) 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package jsonrpc_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/kaspanet/kaspad/jsonrpc"
)

// TestUsageFlagStringer tests the stringized output for the UsageFlag type.
func TestUsageFlagStringer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   jsonrpc.UsageFlag
		want string
	}{
		{0, "0x0"},
		{jsonrpc.UFWebsocketOnly, "UFWebsocketOnly"},
		{jsonrpc.UFNotification, "UFNotification"},
		{jsonrpc.UFWebsocketOnly | jsonrpc.UFNotification,
			"UFWebsocketOnly|UFNotification"},
		{jsonrpc.UFWebsocketOnly | jsonrpc.UFNotification | (1 << 31),
			"UFWebsocketOnly|UFNotification|0x80000000"},
	}

	// Detect additional usage flags that don't have the stringer added.
	numUsageFlags := 0
	highestUsageFlagBit := jsonrpc.TstHighestUsageFlagBit
	for highestUsageFlagBit > 1 {
		numUsageFlags++
		highestUsageFlagBit >>= 1
	}
	if len(tests)-3 != numUsageFlags {
		t.Errorf("It appears a usage flag was added without adding " +
			"an associated stringer test")
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result := test.in.String()
		if result != test.want {
			t.Errorf("String #%d\n got: %s want: %s", i, result,
				test.want)
			continue
		}
	}
}

// TestRegisterCmdErrors ensures the RegisterCmd function returns the expected
// error when provided with invalid types.
func TestRegisterCmdErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		method  string
		cmdFunc func() interface{}
		flags   jsonrpc.UsageFlag
		err     jsonrpc.Error
	}{
		{
			name:   "duplicate method",
			method: "getBlock",
			cmdFunc: func() interface{} {
				return struct{}{}
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrDuplicateMethod},
		},
		{
			name:   "invalid usage flags",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				return 0
			},
			flags: jsonrpc.TstHighestUsageFlagBit,
			err:   jsonrpc.Error{ErrorCode: jsonrpc.ErrInvalidUsageFlags},
		},
		{
			name:   "invalid type",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				return 0
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrInvalidType},
		},
		{
			name:   "invalid type 2",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				return &[]string{}
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrInvalidType},
		},
		{
			name:   "embedded field",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ int }
				return (*test)(nil)
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrEmbeddedType},
		},
		{
			name:   "unexported field",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ a int }
				return (*test)(nil)
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrUnexportedField},
		},
		{
			name:   "unsupported field type 1",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ A **int }
				return (*test)(nil)
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrUnsupportedFieldType},
		},
		{
			name:   "unsupported field type 2",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ A chan int }
				return (*test)(nil)
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrUnsupportedFieldType},
		},
		{
			name:   "unsupported field type 3",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ A complex64 }
				return (*test)(nil)
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrUnsupportedFieldType},
		},
		{
			name:   "unsupported field type 4",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ A complex128 }
				return (*test)(nil)
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrUnsupportedFieldType},
		},
		{
			name:   "unsupported field type 5",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ A func() }
				return (*test)(nil)
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrUnsupportedFieldType},
		},
		{
			name:   "unsupported field type 6",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ A interface{} }
				return (*test)(nil)
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrUnsupportedFieldType},
		},
		{
			name:   "required after optional",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct {
					A *int
					B int
				}
				return (*test)(nil)
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrNonOptionalField},
		},
		{
			name:   "non-optional with default",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct {
					A int `jsonrpcdefault:"1"`
				}
				return (*test)(nil)
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrNonOptionalDefault},
		},
		{
			name:   "mismatched default",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct {
					A *int `jsonrpcdefault:"1.7"`
				}
				return (*test)(nil)
			},
			err: jsonrpc.Error{ErrorCode: jsonrpc.ErrMismatchedDefault},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		err := jsonrpc.RegisterCmd(test.method, test.cmdFunc(),
			test.flags)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("Test #%d (%s) wrong error - got %T, "+
				"want %T", i, test.name, err, test.err)
			continue
		}
		gotErrorCode := err.(jsonrpc.Error).ErrorCode
		if gotErrorCode != test.err.ErrorCode {
			t.Errorf("Test #%d (%s) mismatched error code - got "+
				"%v, want %v", i, test.name, gotErrorCode,
				test.err.ErrorCode)
			continue
		}
	}
}

// TestMustRegisterCmdPanic ensures the MustRegisterCommand function panics when
// used to register an invalid type.
func TestMustRegisterCmdPanic(t *testing.T) {
	t.Parallel()

	// Setup a defer to catch the expected panic to ensure it actually
	// paniced.
	defer func() {
		if err := recover(); err == nil {
			t.Error("MustRegisterCommand did not panic as expected")
		}
	}()

	// Intentionally try to register an invalid type to force a panic.
	jsonrpc.MustRegisterCommand("panicme", 0, 0)
}

// TestRegisteredCmdMethods tests the RegisteredCmdMethods function ensure it
// works as expected.
func TestRegisteredCmdMethods(t *testing.T) {
	t.Parallel()

	// Ensure the registered methods are returned.
	methods := jsonrpc.RegisteredCmdMethods()
	if len(methods) == 0 {
		t.Fatal("RegisteredCmdMethods: no methods")
	}

	// Ensure the returned methods are sorted.
	sortedMethods := make([]string, len(methods))
	copy(sortedMethods, methods)
	sort.Sort(sort.StringSlice(sortedMethods))
	if !reflect.DeepEqual(sortedMethods, methods) {
		t.Fatal("RegisteredCmdMethods: methods are not sorted")
	}
}
