// Copyright (c) 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package model_test

import (
	"github.com/pkg/errors"
	"reflect"
	"sort"
	"testing"

	"github.com/kaspanet/kaspad/rpc/model"
)

// TestUsageFlagStringer tests the stringized output for the UsageFlag type.
func TestUsageFlagStringer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   model.UsageFlag
		want string
	}{
		{0, "0x0"},
		{model.UFWebsocketOnly, "UFWebsocketOnly"},
		{model.UFNotification, "UFNotification"},
		{model.UFWebsocketOnly | model.UFNotification,
			"UFWebsocketOnly|UFNotification"},
		{model.UFWebsocketOnly | model.UFNotification | (1 << 31),
			"UFWebsocketOnly|UFNotification|0x80000000"},
	}

	// Detect additional usage flags that don't have the stringer added.
	numUsageFlags := 0
	highestUsageFlagBit := model.TstHighestUsageFlagBit
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
		flags   model.UsageFlag
		err     model.Error
	}{
		{
			name:   "duplicate method",
			method: "getBlock",
			cmdFunc: func() interface{} {
				return struct{}{}
			},
			err: model.Error{ErrorCode: model.ErrDuplicateMethod},
		},
		{
			name:   "invalid usage flags",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				return 0
			},
			flags: model.TstHighestUsageFlagBit,
			err:   model.Error{ErrorCode: model.ErrInvalidUsageFlags},
		},
		{
			name:   "invalid type",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				return 0
			},
			err: model.Error{ErrorCode: model.ErrInvalidType},
		},
		{
			name:   "invalid type 2",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				return &[]string{}
			},
			err: model.Error{ErrorCode: model.ErrInvalidType},
		},
		{
			name:   "embedded field",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ int }
				return (*test)(nil)
			},
			err: model.Error{ErrorCode: model.ErrEmbeddedType},
		},
		{
			name:   "unexported field",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ a int }
				return (*test)(nil)
			},
			err: model.Error{ErrorCode: model.ErrUnexportedField},
		},
		{
			name:   "unsupported field type 1",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ A **int }
				return (*test)(nil)
			},
			err: model.Error{ErrorCode: model.ErrUnsupportedFieldType},
		},
		{
			name:   "unsupported field type 2",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ A chan int }
				return (*test)(nil)
			},
			err: model.Error{ErrorCode: model.ErrUnsupportedFieldType},
		},
		{
			name:   "unsupported field type 3",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ A complex64 }
				return (*test)(nil)
			},
			err: model.Error{ErrorCode: model.ErrUnsupportedFieldType},
		},
		{
			name:   "unsupported field type 4",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ A complex128 }
				return (*test)(nil)
			},
			err: model.Error{ErrorCode: model.ErrUnsupportedFieldType},
		},
		{
			name:   "unsupported field type 5",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ A func() }
				return (*test)(nil)
			},
			err: model.Error{ErrorCode: model.ErrUnsupportedFieldType},
		},
		{
			name:   "unsupported field type 6",
			method: "registerTestCmd",
			cmdFunc: func() interface{} {
				type test struct{ A interface{} }
				return (*test)(nil)
			},
			err: model.Error{ErrorCode: model.ErrUnsupportedFieldType},
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
			err: model.Error{ErrorCode: model.ErrNonOptionalField},
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
			err: model.Error{ErrorCode: model.ErrNonOptionalDefault},
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
			err: model.Error{ErrorCode: model.ErrMismatchedDefault},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		err := model.RegisterCmd(test.method, test.cmdFunc(),
			test.flags)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("Test #%d (%s) wrong error - got %T, "+
				"want %T", i, test.name, err, test.err)
			continue
		}
		var gotRPCModelErr model.Error
		errors.As(err, &gotRPCModelErr)
		gotErrorCode := gotRPCModelErr.ErrorCode
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
	model.MustRegisterCommand("panicme", 0, 0)
}

// TestRegisteredCmdMethods tests the RegisteredCmdMethods function ensure it
// works as expected.
func TestRegisteredCmdMethods(t *testing.T) {
	t.Parallel()

	// Ensure the registered methods are returned.
	methods := model.RegisteredCmdMethods()
	if len(methods) == 0 {
		t.Fatal("RegisteredCmdMethods: no methods")
	}

	// Ensure the returned methods are sorted.
	sortedMethods := make([]string, len(methods))
	copy(sortedMethods, methods)
	sort.Strings(sortedMethods)
	if !reflect.DeepEqual(sortedMethods, methods) {
		t.Fatal("RegisteredCmdMethods: methods are not sorted")
	}
}
