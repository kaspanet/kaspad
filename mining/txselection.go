package mining

import (
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/random"
	"github.com/daglabs/btcd/util/subnetworkid"
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
	p       float64
	start   float64
	end     float64
	wasUsed bool

	txValue float64

	txMass        uint64
	gasLimit      uint64
	numP2SHSigOps int
}

type txsForBlockTemplate struct {
	selectedTxs   []*util.Tx
	txMasses      []uint64
	txFees        []uint64
	txSigOpCounts []int64
	totalMass     uint64
	totalFees     uint64
	totalSigOps   int64
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
// * Instead of removing a candidate after each iteration, mark it as "used"
//   and only rebalance once rebalanceThreshold * Σ(tx.Value^alpha) of all
//   candidate transactions were marked as "used".
func (g *BlkTmplGenerator) selectTxs(payToAddress util.Address) (*txsForBlockTemplate, error) {
	// Fetch the source transactions.
	sourceTxs := g.txSource.MiningDescs()

	// Create a new txsForBlockTemplate struct, onto which all selectedTxs
	// will be appended.
	txsForBlockTemplate, err := g.newTxsForBlockTemplate(payToAddress, sourceTxs)
	if err != nil {
		return nil, err
	}

	// Collect candidateTxs while excluding txs that will certainly not
	// be selected.
	candidateTxs := g.collectCandidatesTxs(sourceTxs)

	log.Debugf("Considering %d transactions for inclusion to new block",
		len(candidateTxs))

	// Choose which transactions make it into the block.
	g.iterateCandidateTxs(candidateTxs, txsForBlockTemplate)

	return txsForBlockTemplate, nil
}

// newTxsForBlockTemplate creates a txsForBlockTemplate and initializes it
// with a coinbase transaction.
func (g *BlkTmplGenerator) newTxsForBlockTemplate(payToAddress util.Address, sourceTxs []*TxDesc) (*txsForBlockTemplate, error) {
	// Create a new txsForBlockTemplate struct and initialize all the slices
	// to have the max amount of txs, which are the source txs + coinbase.
	// The struct holds the mass, the fees, and number of signature operations
	// for each of the selected transactions and adds an entry for the coinbase.
	// This allows the code below to simply append details about a transaction
	// as it is selected for inclusion in the final block.
	txsForBlockTemplate := &txsForBlockTemplate{
		selectedTxs:   make([]*util.Tx, 0, len(sourceTxs)+1),
		txMasses:      make([]uint64, 0, len(sourceTxs)+1),
		txFees:        make([]uint64, 0, len(sourceTxs)+1),
		txSigOpCounts: make([]int64, 0, len(sourceTxs)+1),
	}

	coinbasePayloadPkScript, err := txscript.PayToAddrScript(payToAddress)
	if err != nil {
		return nil, err
	}
	extraNonce, err := random.Uint64()
	if err != nil {
		return nil, err
	}
	coinbasePayloadExtraData, err := CoinbasePayloadExtraData(extraNonce)
	if err != nil {
		return nil, err
	}
	coinbaseTx, err := g.dag.NextBlockCoinbaseTransactionNoLock(coinbasePayloadPkScript, coinbasePayloadExtraData)
	if err != nil {
		return nil, err
	}
	coinbaseTxMass, err := blockdag.CalcTxMass(coinbaseTx, g.dag.UTXOSet())
	if err != nil {
		return nil, err
	}
	numCoinbaseSigOps := int64(blockdag.CountSigOps(coinbaseTx))

	// Add the coinbase.
	txsForBlockTemplate.selectedTxs = append(txsForBlockTemplate.selectedTxs, coinbaseTx)
	txsForBlockTemplate.totalMass = coinbaseTxMass
	txsForBlockTemplate.totalSigOps = numCoinbaseSigOps
	txsForBlockTemplate.totalFees = uint64(0)
	txsForBlockTemplate.txMasses = append(txsForBlockTemplate.txMasses, coinbaseTxMass)
	txsForBlockTemplate.txFees = append(txsForBlockTemplate.txFees, 0) // For coinbase tx
	txsForBlockTemplate.txSigOpCounts = append(txsForBlockTemplate.txSigOpCounts, numCoinbaseSigOps)

	return txsForBlockTemplate, nil
}

// collectCandidateTxs goes over the sourceTxs and collects only the ones that
// may be included in the next block.
func (g *BlkTmplGenerator) collectCandidatesTxs(sourceTxs []*TxDesc) []*candidateTx {
	nextBlockBlueScore := g.dag.VirtualBlueScore()

	candidateTxs := make([]*candidateTx, 0, len(sourceTxs))
	for _, txDesc := range sourceTxs {
		tx := txDesc.Tx

		// A block can't have more than one coinbase or contain
		// non-finalized transactions.
		if tx.IsCoinBase() {
			log.Warnf("Skipping coinbase tx %s", tx.ID())
			continue
		}
		if !blockdag.IsFinalizedTransaction(tx, nextBlockBlueScore,
			g.timeSource.AdjustedTime()) {
			log.Warnf("Skipping non-finalized tx %s", tx.ID())
			continue
		}

		if txDesc.Fee == 0 {
			log.Warnf("Skipped zero-fee tx %s", tx.ID())
			continue
		}

		txMass, err := blockdag.CalcTxMass(tx, g.dag.UTXOSet())
		if err != nil {
			log.Warnf("Skipping tx %s due to error in "+
				"CalcTxMass: %s", tx.ID(), err)
			continue
		}

		gasLimit := uint64(0)
		if !tx.MsgTx().SubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) && !tx.MsgTx().SubnetworkID.IsBuiltIn() {
			subnetworkID := tx.MsgTx().SubnetworkID
			gasLimit, err = g.dag.SubnetworkStore.GasLimit(&subnetworkID)
			if err != nil {
				log.Warnf("Skipping tx %s due to error in "+
					"GasLimit: %s", tx.ID(), err)
				continue
			}
		}

		numP2SHSigOps, err := blockdag.CountP2SHSigOps(tx, false,
			g.dag.UTXOSet())
		if err != nil {
			log.Warnf("Skipping tx %s due to error in "+
				"GetSigOpCost: %s", tx.ID(), err)
			continue
		}

		// Ensure the transaction inputs pass all of the necessary
		// preconditions before allowing it to be added to the block.
		_, err = blockdag.CheckTransactionInputsAndCalulateFee(tx, nextBlockBlueScore,
			g.dag.UTXOSet(), g.dagParams, false)
		if err != nil {
			log.Warnf("Skipping tx %s due to error in "+
				"CheckTransactionInputs: %s", tx.ID(), err)
			continue
		}
		err = blockdag.ValidateTransactionScripts(tx, g.dag.UTXOSet(),
			txscript.StandardVerifyFlags, g.sigCache)
		if err != nil {
			log.Warnf("Skipping tx %s due to error in "+
				"ValidateTransactionScripts: %s", tx.ID(), err)
			continue
		}

		// Calculate the tx value
		txValue, err := g.calcTxValue(tx, txDesc.Fee)
		if err != nil {
			log.Warnf("Skipping tx %s due to error in "+
				"calcTxValue: %s", tx.ID(), err)
			continue
		}

		candidateTxs = append(candidateTxs, &candidateTx{
			txDesc:        txDesc,
			txValue:       txValue,
			txMass:        txMass,
			gasLimit:      gasLimit,
			numP2SHSigOps: numP2SHSigOps,
		})
	}

	// Sort the candidate txs by their values.
	sort.Slice(candidateTxs, func(i, j int) bool {
		return candidateTxs[i].txValue < candidateTxs[j].txValue
	})

	return candidateTxs
}

// calcTxValue calculates a value to be used in transaction selection.
// The higher the number the more likely it is that the transaction will be
// included in the block.
func (g *BlkTmplGenerator) calcTxValue(tx *util.Tx, fee uint64) (float64, error) {
	mass, err := blockdag.CalcTxMass(tx, g.dag.UTXOSet())
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
	gasLimit, err := g.dag.SubnetworkStore.GasLimit(&msgTx.SubnetworkID)
	if err != nil {
		return 0, err
	}
	return float64(fee) / (float64(mass)/float64(massLimit) + float64(gas)/float64(gasLimit)), nil
}

// iterateCandidateTxs loops over the candidate transactions and appends the
// ones that will be included in the next block into txsForBlockTemplates.
// See selectTxs for further details.
func (g *BlkTmplGenerator) iterateCandidateTxs(candidateTxs []*candidateTx, txsForBlockTemplate *txsForBlockTemplate) {
	usedCount, usedP := 0, 0.0
	candidateTxs, totalP := rebalanceCandidates(candidateTxs, true)
	gasUsageMap := make(map[subnetworkid.SubnetworkID]uint64)

	markCandidateTxUsed := func(candidateTx *candidateTx) {
		candidateTx.wasUsed = true
		usedCount++
		usedP += candidateTx.p
	}

	for len(candidateTxs)-usedCount > 0 {
		// Rebalance the candidates if it's required
		if usedP > 0 && usedP >= rebalanceThreshold*totalP {
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

		// If wasUsed is set, it means we got a collision - ignore and select another Tx
		if selectedTx.wasUsed == true {
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
					if candidateTx.txDesc.Tx.MsgTx().SubnetworkID.IsEqual(&subnetworkID) {
						markCandidateTxUsed(candidateTx)
					}
				}
				continue
			}
			gasUsageMap[subnetworkID] = gasUsage + txGas
		}

		// Enforce maximum signature operations per block. Also check
		// for overflow.
		numSigOps := int64(blockdag.CountSigOps(tx))
		if txsForBlockTemplate.totalSigOps+numSigOps < txsForBlockTemplate.totalSigOps ||
			txsForBlockTemplate.totalSigOps+numSigOps > blockdag.MaxSigOpsPerBlock {
			log.Tracef("Tx %s would exceed the maximum sigops per"+
				"block. As such, stopping.", tx.ID())
			break
		}
		numSigOps += int64(selectedTx.numP2SHSigOps)
		if txsForBlockTemplate.totalSigOps+numSigOps < txsForBlockTemplate.totalSigOps ||
			txsForBlockTemplate.totalSigOps+numSigOps > blockdag.MaxSigOpsPerBlock {
			log.Tracef("Tx %s would exceed the maximum sigops per "+
				"block. As such, stopping.", tx.ID())
			break
		}

		// Add the transaction to the result, increment counters, and
		// save the masses, fees, and signature operation counts to the
		// result.
		txsForBlockTemplate.selectedTxs = append(txsForBlockTemplate.selectedTxs, tx)
		txsForBlockTemplate.totalMass += selectedTx.txMass
		txsForBlockTemplate.totalSigOps += numSigOps
		txsForBlockTemplate.totalFees += selectedTx.txDesc.Fee
		txsForBlockTemplate.txMasses = append(txsForBlockTemplate.txMasses, selectedTx.txMass)
		txsForBlockTemplate.txFees = append(txsForBlockTemplate.txFees, selectedTx.txDesc.Fee)
		txsForBlockTemplate.txSigOpCounts = append(txsForBlockTemplate.txSigOpCounts, numSigOps)

		log.Tracef("Adding tx %s (feePerKB %.2f)",
			tx.ID(), selectedTx.txDesc.FeePerKB)

		markCandidateTxUsed(selectedTx)
	}
}

func rebalanceCandidates(oldCandidateTxs []*candidateTx, isFirstRun bool) (
	candidateTxs []*candidateTx, totalP float64) {

	totalP = 0.0

	candidateTxs = make([]*candidateTx, 0, len(oldCandidateTxs))
	for _, candidateTx := range oldCandidateTxs {
		if candidateTx.wasUsed {
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
