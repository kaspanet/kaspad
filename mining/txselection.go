package mining

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/subnetworkid"
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

type candidateTx struct {
	txDesc  *TxDesc
	txValue float64

	txMass   uint64
	gasLimit uint64

	p     float64
	start float64
	end   float64

	isMarkedForDeletion bool
}

type txsForBlockTemplate struct {
	selectedTxs []*util.Tx
	txMasses    []uint64
	txFees      []uint64
	totalMass   uint64
	totalFees   uint64
}

// selectTxs implements a probabilistic transaction selection algorithm.
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
func (g *BlkTmplGenerator) selectTxs(payToAddress util.Address, extraNonce uint64) (*txsForBlockTemplate, error) {
	// Fetch the source transactions.
	sourceTxs := g.txSource.MiningDescs()

	// Create a new txsForBlockTemplate struct, onto which all selectedTxs
	// will be appended.
	txsForBlockTemplate, err := g.newTxsForBlockTemplate(payToAddress, extraNonce)
	if err != nil {
		return nil, err
	}

	// Collect candidateTxs while excluding txs that will certainly not
	// be selected.
	candidateTxs := g.collectCandidatesTxs(sourceTxs)

	log.Debugf("Considering %d transactions for inclusion to new block",
		len(candidateTxs))

	// Choose which transactions make it into the block.
	g.populateTemplateFromCandidates(candidateTxs, txsForBlockTemplate)

	return txsForBlockTemplate, nil
}

// newTxsForBlockTemplate creates a txsForBlockTemplate and initializes it
// with a coinbase transaction.
func (g *BlkTmplGenerator) newTxsForBlockTemplate(payToAddress util.Address, extraNonce uint64) (*txsForBlockTemplate, error) {
	// Create a new txsForBlockTemplate struct. The struct holds the mass,
	// the fees, and number of signature operations for each of the selected
	// transactions and adds an entry for the coinbase. This allows the code
	// below to simply append details about a transaction as it is selected
	// for inclusion in the final block.
	txsForBlockTemplate := &txsForBlockTemplate{
		selectedTxs: make([]*util.Tx, 0),
		txMasses:    make([]uint64, 0),
		txFees:      make([]uint64, 0),
	}

	coinbasePayloadExtraData, err := blockdag.CoinbasePayloadExtraData(extraNonce, CoinbaseFlags)
	if err != nil {
		return nil, err
	}
	coinbaseTx, err := g.dag.NextCoinbaseFromAddress(payToAddress, coinbasePayloadExtraData)
	if err != nil {
		return nil, err
	}
	coinbaseTxMass, err := blockdag.CalcTxMassFromUTXOSet(coinbaseTx, g.dag.UTXOSet())
	if err != nil {
		return nil, err
	}

	// Add the coinbase.
	txsForBlockTemplate.selectedTxs = append(txsForBlockTemplate.selectedTxs, coinbaseTx)
	txsForBlockTemplate.totalMass = coinbaseTxMass
	txsForBlockTemplate.totalFees = uint64(0)
	txsForBlockTemplate.txMasses = append(txsForBlockTemplate.txMasses, coinbaseTxMass)
	txsForBlockTemplate.txFees = append(txsForBlockTemplate.txFees, 0) // For coinbase tx

	return txsForBlockTemplate, nil
}

// collectCandidateTxs goes over the sourceTxs and collects only the ones that
// may be included in the next block.
func (g *BlkTmplGenerator) collectCandidatesTxs(sourceTxs []*TxDesc) []*candidateTx {
	nextBlockBlueScore := g.dag.VirtualBlueScore()

	candidateTxs := make([]*candidateTx, 0, len(sourceTxs))
	for _, txDesc := range sourceTxs {
		tx := txDesc.Tx

		// A block can't contain non-finalized transactions.
		if !blockdag.IsFinalizedTransaction(tx, nextBlockBlueScore,
			g.timeSource.Now()) {
			log.Debugf("Skipping non-finalized tx %s", tx.ID())
			continue
		}

		// A block can't contain zero-fee transactions.
		if txDesc.Fee == 0 {
			log.Warnf("Skipped zero-fee tx %s", tx.ID())
			continue
		}

		txMass, err := blockdag.CalcTxMassFromUTXOSet(tx, g.dag.UTXOSet())
		if err != nil {
			log.Warnf("Skipping tx %s due to error in "+
				"CalcTxMass: %s", tx.ID(), err)
			continue
		}

		gasLimit := uint64(0)
		if !tx.MsgTx().SubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) && !tx.MsgTx().SubnetworkID.IsBuiltIn() {
			subnetworkID := tx.MsgTx().SubnetworkID
			gasLimit, err = blockdag.GasLimit(&subnetworkID)
			if err != nil {
				log.Warnf("Skipping tx %s due to error in "+
					"GasLimit: %s", tx.ID(), err)
				continue
			}
		}

		// Calculate the tx value
		txValue, err := g.calcTxValue(tx, txDesc.Fee)
		if err != nil {
			log.Warnf("Skipping tx %s due to error in "+
				"calcTxValue: %s", tx.ID(), err)
			continue
		}

		candidateTxs = append(candidateTxs, &candidateTx{
			txDesc:   txDesc,
			txValue:  txValue,
			txMass:   txMass,
			gasLimit: gasLimit,
		})
	}

	// Sort the candidate txs by subnetworkID.
	sort.Slice(candidateTxs, func(i, j int) bool {
		return subnetworkid.Less(&candidateTxs[i].txDesc.Tx.MsgTx().SubnetworkID,
			&candidateTxs[j].txDesc.Tx.MsgTx().SubnetworkID)
	})

	return candidateTxs
}

// calcTxValue calculates a value to be used in transaction selection.
// The higher the number the more likely it is that the transaction will be
// included in the block.
func (g *BlkTmplGenerator) calcTxValue(tx *util.Tx, fee uint64) (float64, error) {
	mass, err := blockdag.CalcTxMassFromUTXOSet(tx, g.dag.UTXOSet())
	if err != nil {
		return 0, err
	}
	massLimit := g.policy.BlockMaxMass

	msgTx := tx.MsgTx()
	if msgTx.SubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) ||
		msgTx.SubnetworkID.IsBuiltIn() {
		return float64(fee) / (float64(mass) / float64(massLimit)), nil
	}

	gas := msgTx.Gas
	gasLimit, err := blockdag.GasLimit(&msgTx.SubnetworkID)
	if err != nil {
		return 0, err
	}
	return float64(fee) / (float64(mass)/float64(massLimit) + float64(gas)/float64(gasLimit)), nil
}

// populateTemplateFromCandidates loops over the candidate transactions
// and appends the ones that will be included in the next block into
// txsForBlockTemplates.
// See selectTxs for further details.
func (g *BlkTmplGenerator) populateTemplateFromCandidates(candidateTxs []*candidateTx, txsForBlockTemplate *txsForBlockTemplate) {
	usedCount, usedP := 0, 0.0
	candidateTxs, totalP := rebalanceCandidates(candidateTxs, true)
	gasUsageMap := make(map[subnetworkid.SubnetworkID]uint64)

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
		tx := selectedTx.txDesc.Tx

		// Enforce maximum transaction mass per block. Also check
		// for overflow.
		if txsForBlockTemplate.totalMass+selectedTx.txMass < txsForBlockTemplate.totalMass ||
			txsForBlockTemplate.totalMass+selectedTx.txMass > g.policy.BlockMaxMass {
			log.Tracef("Tx %s would exceed the max block mass. "+
				"As such, stopping.", tx.ID())
			break
		}

		// Enforce maximum gas per subnetwork per block. Also check
		// for overflow.
		if !tx.MsgTx().SubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) && !tx.MsgTx().SubnetworkID.IsBuiltIn() {
			subnetworkID := tx.MsgTx().SubnetworkID
			gasUsage, ok := gasUsageMap[subnetworkID]
			if !ok {
				gasUsage = 0
			}
			txGas := tx.MsgTx().Gas
			if gasUsage+txGas < gasUsage ||
				gasUsage+txGas > selectedTx.gasLimit {
				log.Tracef("Tx %s would exceed the gas limit in "+
					"subnetwork %s. Removing all remaining txs from this "+
					"subnetwork.",
					tx.MsgTx().TxID(), subnetworkID)
				for _, candidateTx := range candidateTxs {
					// candidateTxs are ordered by subnetwork, so we can safely assume
					// that transactions after subnetworkID will not be relevant.
					if subnetworkid.Less(&subnetworkID, &candidateTx.txDesc.Tx.MsgTx().SubnetworkID) {
						break
					}

					if candidateTx.txDesc.Tx.MsgTx().SubnetworkID.IsEqual(&subnetworkID) {
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
		txsForBlockTemplate.totalMass += selectedTx.txMass
		txsForBlockTemplate.totalFees += selectedTx.txDesc.Fee

		log.Tracef("Adding tx %s (feePerKB %d)",
			tx.ID(), selectedTx.txDesc.FeePerKB)

		markCandidateTxForDeletion(selectedTx)
	}

	sort.Slice(selectedTxs, func(i, j int) bool {
		return subnetworkid.Less(&selectedTxs[i].txDesc.Tx.MsgTx().SubnetworkID,
			&selectedTxs[j].txDesc.Tx.MsgTx().SubnetworkID)
	})
	for _, selectedTx := range selectedTxs {
		txsForBlockTemplate.selectedTxs = append(txsForBlockTemplate.selectedTxs, selectedTx.txDesc.Tx)
		txsForBlockTemplate.txMasses = append(txsForBlockTemplate.txMasses, selectedTx.txMass)
		txsForBlockTemplate.txFees = append(txsForBlockTemplate.txFees, selectedTx.txDesc.Fee)
	}
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
