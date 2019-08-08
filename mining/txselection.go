package mining

import (
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/random"
	"github.com/daglabs/btcd/util/subnetworkid"
	"math"
	"math/rand"
)

const (
	// alpha is used when determining a candidate transaction's
	// initial p value.
	alpha = 3

	// rebalanceThreshold is a percentage of the candidate transaction
	// collection under which we don't rebalance. See selectTxs for details.
	rebalanceThreshold = 0.95
)

type candidateTx struct {
	txDesc  *TxDesc
	p       float64
	start   float64
	end     float64
	wasUsed bool

	txMass        uint64
	gasLimit      uint64
	numP2SHSigOps int
}

type txsForBlockTemplate struct {
	selectedTxs   []*util.Tx
	txMasses      []uint64
	txFees        []uint64
	txSigOpCounts []int64
	blockMass     uint64
	totalFees     uint64
	blockSigOps   int64
}

// selectTxs implements a probabilistic transaction selection algorithm.
// The algorithm, roughly, is as follows:
// 1. We assign a probability to each transaction equal to:
//    (candidateTx.Value^alpha) / Σ(tx.Value^alpha)
//    Where the sum in the denominator is equal to 1.
// 2. We draw a random number in [0,1) and select a transaction accordingly.
// 3. If it's valid, add it to the result and remove it from the candidates.
// 4. Continue iterating the above until we have either selected all
//    available transactions or ran out of gas/block space.
//
// Note that we make two optimizations here:
// * Draw a number in [0,Σ(tx.Value^alpha)) to avoid normalization
// * Instead of removing a candidate after each iteration, mark it as "used"
//   and only rebalance once rebalanceThreshold * Σ(tx.Value^alpha) of all
//   candidate transactions were marked as "used".
func (g *BlkTmplGenerator) selectTxs(payToAddress util.Address) (*txsForBlockTemplate, error) {
	// Fetch the source transactions. We expect here that the transactions
	// have previously been sorted by selection value.
	sourceTxns := g.txSource.MiningDescs()

	// Create the result object and initialize all the slices to have
	// the max amount of txs, which are the source tx + coinbase.
	// The result object holds the mass, the fees, and number of signature
	// operations for each of the selected transactions and adds an entry for
	// the coinbase.  This allows the code below to simply append details
	// about a transaction as it is selected for inclusion in the final block.
	result := &txsForBlockTemplate{
		selectedTxs:   make([]*util.Tx, 0, len(sourceTxns)+1),
		txMasses:      make([]uint64, 0, len(sourceTxns)+1),
		txFees:        make([]uint64, 0, len(sourceTxns)+1),
		txSigOpCounts: make([]int64, 0, len(sourceTxns)+1),
	}

	nextBlockUTXO := g.dag.UTXOSet()
	nextBlockBlueScore := g.dag.VirtualBlueScore()

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
	coinbaseTxMass, err := blockdag.CalcTxMass(coinbaseTx, nextBlockUTXO)
	if err != nil {
		return nil, err
	}
	numCoinbaseSigOps := int64(blockdag.CountSigOps(coinbaseTx))

	// Add the coinbase to the result object. Note that since the total fees
	// aren't known yet, we use a dummy value for the coinbase fee which will
	// be updated later.
	result.selectedTxs = append(result.selectedTxs, coinbaseTx)
	result.blockMass = coinbaseTxMass
	result.blockSigOps = numCoinbaseSigOps
	result.totalFees = uint64(0)
	result.txMasses = append(result.txMasses, coinbaseTxMass)
	result.txFees = append(result.txFees, 0) // For coinbase tx
	result.txSigOpCounts = append(result.txSigOpCounts, numCoinbaseSigOps)

	// Collect candidateTxs while excluding txs that will certainly not
	// be selected.
	candidateTxs := make([]*candidateTx, 0, len(sourceTxns))
	for _, txDesc := range sourceTxns {
		tx := txDesc.Tx

		// A block can't have more than one coinbase or contain
		// non-finalized transactions.
		if tx.IsCoinBase() {
			log.Tracef("Skipping coinbase tx %s", tx.ID())
			continue
		}
		if !blockdag.IsFinalizedTransaction(tx, nextBlockBlueScore,
			g.timeSource.AdjustedTime()) {
			log.Tracef("Skipping non-finalized tx %s", tx.ID())
			continue
		}

		txMass, err := blockdag.CalcTxMass(tx, g.dag.UTXOSet())
		if err != nil {
			log.Tracef("Skipping tx %s due to error in "+
				"CalcTxMass: %s", tx.ID(), err)
			continue
		}

		gasLimit := uint64(0)
		if !tx.MsgTx().SubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) && !tx.MsgTx().SubnetworkID.IsBuiltIn() {
			subnetworkID := tx.MsgTx().SubnetworkID
			gasLimit, err = g.dag.SubnetworkStore.GasLimit(&subnetworkID)
			if err != nil {
				log.Errorf("Cannot get GAS limit for subnetwork %s", subnetworkID)
				continue
			}
		}

		numP2SHSigOps, err := blockdag.CountP2SHSigOps(tx, false,
			g.dag.UTXOSet())
		if err != nil {
			log.Tracef("Skipping tx %s due to error in "+
				"GetSigOpCost: %s", tx.ID(), err)
			continue
		}

		// Ensure the transaction inputs pass all of the necessary
		// preconditions before allowing it to be added to the block.
		_, err = blockdag.CheckTransactionInputsAndCalulateFee(tx, nextBlockBlueScore,
			g.dag.UTXOSet(), g.dagParams, false)
		if err != nil {
			log.Tracef("Skipping tx %s due to error in "+
				"CheckTransactionInputs: %s", tx.ID(), err)
			continue
		}
		err = blockdag.ValidateTransactionScripts(tx, g.dag.UTXOSet(),
			txscript.StandardVerifyFlags, g.sigCache)
		if err != nil {
			log.Tracef("Skipping tx %s due to error in "+
				"ValidateTransactionScripts: %s", tx.ID(), err)
			continue
		}

		candidateTxs = append(candidateTxs, &candidateTx{
			txDesc:        txDesc,
			txMass:        txMass,
			gasLimit:      gasLimit,
			numP2SHSigOps: numP2SHSigOps,
		})
	}

	usedCount, usedP := 0, 0.0
	candidateTxs, totalP := rebalanceCandidates(candidateTxs, usedCount, true)
	gasUsageMap := make(map[subnetworkid.SubnetworkID]uint64)

	markCandidateTxUsed := func(candidateTx *candidateTx) {
		candidateTx.wasUsed = true
		usedCount++
		usedP += candidateTx.p
	}

	rebalanceIfRequired := func() {
		if usedP >= rebalanceThreshold*totalP {
			candidateTxs, totalP = rebalanceCandidates(candidateTxs, usedCount, false)
			usedCount, usedP = 0, 0.0
		}
	}

	log.Debugf("Considering %d transactions for inclusion to new block",
		len(candidateTxs))

	// Choose which transactions make it into the block.
	for len(candidateTxs) > 0 {
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
		if result.blockMass+selectedTx.txMass < result.blockMass ||
			result.blockMass+selectedTx.txMass >= g.policy.BlockMaxMass {
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
			if gasUsage+txGas > selectedTx.gasLimit {
				log.Tracef("Tx %s would exceed the gas limit in "+
					"subnetwork %s. Removing all remaining txs from this "+
					"subnetwork.",
					tx.MsgTx().TxID(), subnetworkID)
				for _, candidateTx := range candidateTxs {
					if candidateTx.txDesc.Tx.MsgTx().SubnetworkID.IsEqual(&subnetworkID) {
						markCandidateTxUsed(candidateTx)
					}
				}
				rebalanceIfRequired()
				continue
			}
			gasUsageMap[subnetworkID] = gasUsage + txGas
		}

		// Enforce maximum signature operations per block. Also check
		// for overflow.
		numSigOps := int64(blockdag.CountSigOps(tx))
		if result.blockSigOps+numSigOps < result.blockSigOps ||
			result.blockSigOps+numSigOps > blockdag.MaxSigOpsPerBlock {
			log.Tracef("Tx %s would exceed the maximum sigops per"+
				"block. As such, stopping.", tx.ID())
			break
		}
		numSigOps += int64(selectedTx.numP2SHSigOps)
		if result.blockSigOps+numSigOps < result.blockSigOps ||
			result.blockSigOps+numSigOps > blockdag.MaxSigOpsPerBlock {
			log.Tracef("Tx %s would exceed the maximum sigops per "+
				"block. As such, stopping.", tx.ID())
			break
		}

		// Add the transaction to the result, increment counters, and
		// save the masses, fees, and signature operation counts to the
		// result.
		result.selectedTxs = append(result.selectedTxs, tx)
		result.blockMass += selectedTx.txMass
		result.blockSigOps += numSigOps
		result.totalFees += selectedTx.txDesc.Fee
		result.txMasses = append(result.txMasses, selectedTx.txMass)
		result.txFees = append(result.txFees, selectedTx.txDesc.Fee)
		result.txSigOpCounts = append(result.txSigOpCounts, numSigOps)

		log.Tracef("Adding tx %s (feePerKB %.2f)",
			tx.ID(), selectedTx.txDesc.FeePerKB)

		markCandidateTxUsed(selectedTx)
		rebalanceIfRequired()
	}

	return result, nil
}

func rebalanceCandidates(oldCandidateTxs []*candidateTx, usedCount int, isFirstRun bool) (
	candidateTxs []*candidateTx, totalP float64) {

	candidateTxs = make([]*candidateTx, 0, len(oldCandidateTxs)-usedCount)

	for _, candidateTx := range oldCandidateTxs {
		if candidateTx.wasUsed {
			continue
		}

		candidateTxs = append(candidateTxs, candidateTx)
	}

	totalP = 0.0

	for _, candidateTx := range candidateTxs {
		if isFirstRun {
			candidateTx.p = math.Pow(candidateTx.txDesc.SelectionValue, alpha)
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
