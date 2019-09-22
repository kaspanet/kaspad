package rpc

import (
	"fmt"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/util/daghash"
	"strings"
)

// handleGetBlockDAGInfo implements the getBlockDagInfo command.
func handleGetBlockDAGInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Obtain a snapshot of the current best known DAG state. We'll
	// populate the response to this call primarily from this snapshot.
	params := s.cfg.DAGParams
	dag := s.cfg.DAG

	dagInfo := &btcjson.GetBlockDAGInfoResult{
		DAG:            params.Name,
		Blocks:         dag.BlockCount(),
		Headers:        dag.BlockCount(),
		TipHashes:      daghash.Strings(dag.TipHashes()),
		Difficulty:     getDifficultyRatio(dag.CurrentBits(), params),
		MedianTime:     dag.CalcPastMedianTime().Unix(),
		UTXOCommitment: dag.UTXOCommitment(),
		Pruned:         false,
		Bip9SoftForks:  make(map[string]*btcjson.Bip9SoftForkDescription),
	}

	// Finally, query the BIP0009 version bits state for all currently
	// defined BIP0009 soft-fork deployments.
	for deployment, deploymentDetails := range params.Deployments {
		// Map the integer deployment ID into a human readable
		// fork-name.
		var forkName string
		switch deployment {
		case dagconfig.DeploymentTestDummy:
			forkName = "dummy"

		default:
			return nil, &btcjson.RPCError{
				Code: btcjson.ErrRPCInternal.Code,
				Message: fmt.Sprintf("Unknown deployment %d "+
					"detected", deployment),
			}
		}

		// Query the dag for the current status of the deployment as
		// identified by its deployment ID.
		deploymentStatus, err := dag.ThresholdState(uint32(deployment))
		if err != nil {
			context := "Failed to obtain deployment status"
			return nil, internalRPCError(err.Error(), context)
		}

		// Attempt to convert the current deployment status into a
		// human readable string. If the status is unrecognized, then a
		// non-nil error is returned.
		statusString, err := softForkStatus(deploymentStatus)
		if err != nil {
			return nil, &btcjson.RPCError{
				Code: btcjson.ErrRPCInternal.Code,
				Message: fmt.Sprintf("unknown deployment status: %d",
					deploymentStatus),
			}
		}

		// Finally, populate the soft-fork description with all the
		// information gathered above.
		dagInfo.Bip9SoftForks[forkName] = &btcjson.Bip9SoftForkDescription{
			Status:    strings.ToLower(statusString),
			Bit:       deploymentDetails.BitNumber,
			StartTime: int64(deploymentDetails.StartTime),
			Timeout:   int64(deploymentDetails.ExpireTime),
		}
	}

	return dagInfo, nil
}
