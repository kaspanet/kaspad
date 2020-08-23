package dagconfig_test

import (
	"testing"

	. "github.com/kaspanet/kaspad/domain/dagconfig"
)

// Define some of the required parameters for a user-registered
// network. This is necessary to test the registration of and
// lookup of encoding magics from the network.
var mockNetParams = Params{
	Name: "mocknet",
	Net:  1<<32 - 1,
}

func TestRegister(t *testing.T) {
	type registerTest struct {
		name   string
		params *Params
		err    error
	}

	tests := []struct {
		name     string
		register []registerTest
	}{
		{
			name: "default networks",
			register: []registerTest{
				{
					name:   "duplicate mainnet",
					params: &MainnetParams,
					err:    ErrDuplicateNet,
				},
				{
					name:   "duplicate testnet",
					params: &TestnetParams,
					err:    ErrDuplicateNet,
				},
				{
					name:   "duplicate simnet",
					params: &SimnetParams,
					err:    ErrDuplicateNet,
				},
			},
		},
		{
			name: "register mocknet",
			register: []registerTest{
				{
					name:   "mocknet",
					params: &mockNetParams,
					err:    nil,
				},
			},
		},
		{
			name: "more duplicates",
			register: []registerTest{
				{
					name:   "duplicate mainnet",
					params: &MainnetParams,
					err:    ErrDuplicateNet,
				},
				{
					name:   "duplicate testnet",
					params: &TestnetParams,
					err:    ErrDuplicateNet,
				},
				{
					name:   "duplicate simnet",
					params: &SimnetParams,
					err:    ErrDuplicateNet,
				},
				{
					name:   "duplicate mocknet",
					params: &mockNetParams,
					err:    ErrDuplicateNet,
				},
			},
		},
	}

	for _, test := range tests {
		for _, network := range test.register {
			err := Register(network.params)

			if err != network.err {
				t.Errorf("%s:%s: Registered network with unexpected error: got %v expected %v",
					network.name, network.name, err, network.err)
			}
		}
	}
}
