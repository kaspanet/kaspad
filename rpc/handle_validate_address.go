package rpc

import (
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util"
)

// handleValidateAddress implements the validateAddress command.
func handleValidateAddress(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*rpcmodel.ValidateAddressCmd)

	result := rpcmodel.ValidateAddressResult{}
	addr, err := util.DecodeAddress(c.Address, s.dag.Params.Prefix)
	if err != nil {
		// Return the default value (false) for IsValid.
		return result, nil
	}

	result.Address = addr.EncodeAddress()
	result.IsValid = true

	return result, nil
}
