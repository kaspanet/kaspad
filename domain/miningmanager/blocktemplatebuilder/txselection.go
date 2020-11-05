package blocktemplatebuilder

import (
	consensusexternalapi "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"math"
	"math/rand"
	"sort"
)

const (
	// alpha is a coefficient that defines how uniform the distribution of
	// candidate transactions should be. A smaller alpha makes the distribution
	// more uniform. Alpha is used when determining a candidate transaction's
	// initial p value.
	alpha = 3

	// rebalanceThreshold is the percentage of candidate transactions under which
	// we don't rebalance. Rebalancing is a heavy operation so we prefer to avoid
	// rebalancing very often. On the other hand, if we don't rebalance often enough
	// we risk having too many collisions.
	// The value is derived from the max probability of collision. That is to say,
	// if rebalanceThreshold is 0.95, there's a 1-in-20 chance of collision.
	// See selectTxs for further details.
	rebalanceThreshold = 0.95
)

type selectedTransactions struct {
	selectedTxs []*consensusexternalapi.DomainTransaction
	txMasses    []uint64
	txFees      []uint64
	totalMass   uint64
	totalFees   uint64
}

// selectTransactions implements a probabilistic transaction selection algorithm.
// The algorithm, roughly, is as follows:
// 1. We assign a probability to each transaction equal to:
//    (candidateTx.Value^alpha) / Σ(tx.Value^alpha)
//    Where the sum of the probabilities of all txs is 1.
// 2. We draw a random number in [0,1) and select a transaction accordingly.
// 3. If it's valid, add it to the selectedTxs and remove it from the candidates.
// 4. Continue iterating the above until we have either selected all
//    available transactions or ran out of gas/block space.
//
// Note that we make two optimizations here:
// * Draw a number in [0,Σ(tx.Value^alpha)) to avoid normalization
// * Instead of removing a candidate after each iteration, mark it for deletion.
//   Once the sum of probabilities of marked transactions is greater than
//   rebalanceThreshold percent of the sum of probabilities of all transactions,
//   rebalance.

// selectTransactions loops over the candidate transactions
// and appends the ones that will be included in the next block into
// txsForBlockTemplates.
// See selectTxs for further details.
func (btb *blockTemplateBuilder) selectTransactions(candidateTxs []*candidateTx) selectedTransactions {
	txsForBlockTemplate := selectedTransactions{
		selectedTxs: make([]*consensusexternalapi.DomainTransaction, 0, len(candidateTxs)),
		txMasses:    make([]uint64, 0, len(candidateTxs)),
		txFees:      make([]uint64, 0, len(candidateTxs)),
		totalMass:   0,
		totalFees:   0,
	}
	usedCount, usedP := 0, 0.0
	candidateTxs, totalP := rebalanceCandidates(candidateTxs, true)
	gasUsageMap := make(map[consensusexternalapi.DomainSubnetworkID]uint64)

	markCandidateTxForDeletion := func(candidateTx *candidateTx) {
		candidateTx.isMarkedForDeletion = true
		usedCount++
		usedP += candidateTx.p
	}

	selectedTxs := make([]*candidateTx, 0)
	for len(candidateTxs)-usedCount > 0 {
		// Rebalance the candidates if it's required
		if usedP >= rebalanceThreshold*totalP {
			candidateTxs, totalP = rebalanceCandidates(candidateTxs, false)
			usedCount, usedP = 0, 0.0

			// Break if we now ran out of transactions
			if len(candidateTxs) == 0 {
				break
			}
		}

		// Select a candidate tx at random
		r := rand.Float64()
		r *= totalP
		selectedTx := findTx(candidateTxs, r)

		// If isMarkedForDeletion is set, it means we got a collision.
		// Ignore and select another Tx.
		if selectedTx.isMarkedForDeletion {
			continue
		}
		tx := selectedTx.DomainTransaction

		// Enforce maximum transaction mass per block. Also check
		// for overflow.
		if txsForBlockTemplate.totalMass+selectedTx.Mass < txsForBlockTemplate.totalMass ||
			txsForBlockTemplate.totalMass+selectedTx.Mass > btb.policy.BlockMaxMass {
			log.Tracef("Tx %s would exceed the max block mass. "+
				"As such, stopping.", consensusserialization.TransactionID(tx))
			break
		}

		// Enforce maximum gas per subnetwork per block. Also check
		// for overflow.
		if !subnetworks.IsBuiltInOrNative(tx.SubnetworkID) {
			subnetworkID := tx.SubnetworkID
			gasUsage, ok := gasUsageMap[subnetworkID]
			if !ok {
				gasUsage = 0
			}
			txGas := tx.Gas
			if gasUsage+txGas < gasUsage ||
				gasUsage+txGas > selectedTx.gasLimit {
				log.Tracef("Tx %s would exceed the gas limit in "+
					"subnetwork %s. Removing all remaining txs from this "+
					"subnetwork.",
					consensusserialization.TransactionID(tx), subnetworkID)
				for _, candidateTx := range candidateTxs {
					// candidateTxs are ordered by subnetwork, so we can safely assume
					// that transactions after subnetworkID will not be relevant.
					if subnetworks.Less(subnetworkID, candidateTx.SubnetworkID) {
						break
					}

					if candidateTx.SubnetworkID == subnetworkID {
						markCandidateTxForDeletion(candidateTx)
					}
				}
				continue
			}
			gasUsageMap[subnetworkID] = gasUsage + txGas
		}

		// Add the transaction to the result, increment counters, and
		// save the masses, fees, and signature operation counts to the
		// result.
		selectedTxs = append(selectedTxs, selectedTx)
		txsForBlockTemplate.totalMass += selectedTx.Mass
		txsForBlockTemplate.totalFees += selectedTx.Fee

		log.Tracef("Adding tx %s (feePerMegaGram %d)",
			consensusserialization.TransactionID(tx), selectedTx.Fee*1e6/selectedTx.Mass)

		markCandidateTxForDeletion(selectedTx)
	}

	sort.Slice(selectedTxs, func(i, j int) bool {
		return subnetworks.Less(selectedTxs[i].SubnetworkID, selectedTxs[j].SubnetworkID)
	})
	for _, selectedTx := range selectedTxs {
		txsForBlockTemplate.selectedTxs = append(txsForBlockTemplate.selectedTxs, selectedTx.DomainTransaction)
		txsForBlockTemplate.txMasses = append(txsForBlockTemplate.txMasses, selectedTx.Mass)
		txsForBlockTemplate.txFees = append(txsForBlockTemplate.txFees, selectedTx.Fee)
	}
	return txsForBlockTemplate
}

func rebalanceCandidates(oldCandidateTxs []*candidateTx, isFirstRun bool) (
	candidateTxs []*candidateTx, totalP float64) {

	totalP = 0.0

	candidateTxs = make([]*candidateTx, 0, len(oldCandidateTxs))
	for _, candidateTx := range oldCandidateTxs {
		if candidateTx.isMarkedForDeletion {
			continue
		}

		candidateTxs = append(candidateTxs, candidateTx)
	}

	for _, candidateTx := range candidateTxs {
		if isFirstRun {
			candidateTx.p = math.Pow(candidateTx.txValue, alpha)
		}
		candidateTx.start = totalP
		candidateTx.end = totalP + candidateTx.p

		totalP += candidateTx.p
	}
	return
}

// findTx finds the candidateTx in whose range r falls.
// For example, if we have candidateTxs with starts and ends:
// * tx1: start 0,   end 100
// * tx2: start 100, end 105
// * tx3: start 105, end 2000
// And r=102, then findTx will return tx2.
func findTx(candidateTxs []*candidateTx, r float64) *candidateTx {
	min := 0
	max := len(candidateTxs) - 1
	for {
		i := (min + max) / 2
		candidateTx := candidateTxs[i]
		if candidateTx.end < r {
			min = i + 1
			continue
		} else if candidateTx.start > r {
			max = i - 1
			continue
		}
		return candidateTx
	}
}
