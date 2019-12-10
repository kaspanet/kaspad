package controllers

import "github.com/kaspanet/kaspad/kasparov/kasparovserver/apimodels"

// GetFeeEstimatesHandler returns the fee estimates for different priorities
// for accepting a transaction in the DAG.
func GetFeeEstimatesHandler() (interface{}, error) {
	return &apimodels.FeeEstimateResponse{
		HighPriority:   3,
		NormalPriority: 2,
		LowPriority:    1,
	}, nil
}
