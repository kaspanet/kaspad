package controllers

import (
	"github.com/daglabs/btcd/kasparov/server/models"
)

// GetFeeEstimatesHandler returns the fee estimates for different priorities
// for accepting a transaction in the DAG.
func GetFeeEstimatesHandler() (interface{}, error) {
	return &models.FeeEstimateResponse{
		HighPriority:   3,
		NormalPriority: 2,
		LowPriority:    1,
	}, nil
}
