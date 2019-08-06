package mining

import (
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/random"
	"github.com/daglabs/btcd/util/subnetworkid"
)

type txsForBlockTemplate struct {
	selectedTxs   []*util.Tx
	txMasses      []uint64
	txFees        []uint64
	txSigOpCounts []int64
	blockMass     uint64
	totalFees     uint64
	blockSigOps   int64
}

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

	// Create map of GAS usage per subnetwork
	gasUsageMap := make(map[subnetworkid.SubnetworkID]uint64)

	log.Debugf("Considering %d transactions for inclusion to new block",
		len(sourceTxns))

	// Choose which transactions make it into the block.
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

		if !tx.MsgTx().SubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) && !tx.MsgTx().SubnetworkID.IsBuiltIn() {
			subnetworkID := tx.MsgTx().SubnetworkID
			gasUsage, ok := gasUsageMap[subnetworkID]
			if !ok {
				gasUsage = 0
			}
			gasLimit, err := g.dag.SubnetworkStore.GasLimit(&subnetworkID)
			if err != nil {
				log.Errorf("Cannot get GAS limit for subnetwork %s", subnetworkID)
				continue
			}
			txGas := tx.MsgTx().Gas
			if gasLimit-gasUsage < txGas {
				log.Tracef("Transaction %s (GAS=%d) ignored because gas overusage (GASUsage=%d) in subnetwork %s (GASLimit=%d)",
					tx.MsgTx().TxID(), txGas, gasUsage, subnetworkID, gasLimit)
				continue
			}
			gasUsageMap[subnetworkID] = gasUsage + txGas
		}

		// Enforce maximum transaction mass per block. Also check
		// for overflow.
		txMass, err := blockdag.CalcTxMass(tx, g.dag.UTXOSet())
		if err != nil {
			log.Tracef("Skipping tx %s due to error in "+
				"CalcTxMass: %s", tx.ID(), err)
			continue
		}
		if result.blockMass+txMass < result.blockMass ||
			result.blockMass >= g.policy.BlockMaxMass {
			log.Tracef("Skipping tx %s because it would exceed "+
				"the max block mass", tx.ID())
			continue
		}

		// Enforce maximum signature operations per block. Also check
		// for overflow.
		numSigOps := int64(blockdag.CountSigOps(tx))
		if result.blockSigOps+numSigOps < result.blockSigOps ||
			result.blockSigOps+numSigOps > blockdag.MaxSigOpsPerBlock {
			log.Tracef("Skipping tx %s because it would exceed "+
				"the maximum sigops per block", tx.ID())
			continue
		}
		numP2SHSigOps, err := blockdag.CountP2SHSigOps(tx, false,
			g.dag.UTXOSet())
		if err != nil {
			log.Tracef("Skipping tx %s due to error in "+
				"GetSigOpCost: %s", tx.ID(), err)
			continue
		}
		numSigOps += int64(numP2SHSigOps)
		if result.blockSigOps+numSigOps < result.blockSigOps ||
			result.blockSigOps+numSigOps > blockdag.MaxSigOpsPerBlock {
			log.Tracef("Skipping tx %s because it would "+
				"exceed the maximum sigops per block", tx.ID())
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

		// Add the transaction to the block, increment counters, and
		// save the masses, fees, and signature operation counts to the block
		// template.
		result.selectedTxs = append(result.selectedTxs, tx)
		result.blockMass += txMass
		result.blockSigOps += numSigOps
		result.totalFees += txDesc.Fee
		result.txMasses = append(result.txMasses, txMass)
		result.txFees = append(result.txFees, txDesc.Fee)
		result.txSigOpCounts = append(result.txSigOpCounts, numSigOps)

		log.Tracef("Adding tx %s (feePerKB %.2f)",
			tx.ID(), txDesc.FeePerKB)
	}

	return result, nil
}
