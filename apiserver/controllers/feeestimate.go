package controllers

import (
	"github.com/daglabs/btcd/apiserver/apimodels"
	"github.com/daglabs/btcd/httpserverutils"
)

// GetFeeEstimatesHandler returns the fee estimates for different priorities
// for accepting a transaction in the DAG.
func GetFeeEstimatesHandler() (interface{}, *httpserverutils.HandlerError) {
	return &apimodels.FeeEstimateResponse{
		HighPriority:   3,
		NormalPriority: 2,
		LowPriority:    1,
	}, nil
}
