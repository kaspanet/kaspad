package main

import (
	"bytes"
	"encoding/hex"
	"github.com/daglabs/btcd/apiserver/models"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/jinzhu/gorm"
	"strconv"
	"time"
)

func blockLoop(client *apiServerClient, db *gorm.DB, doneChan chan struct{}) error {
	mostRecentBlockHash := findMostRecentBlockHash(db)
	err := SyncBlocks(client, db, mostRecentBlockHash)
	if err != nil {
		return err
	}

	err = SyncSelectedParentChain(client, db, mostRecentBlockHash)
	if err != nil {
		return err
	}

loop:
	for {
		select {
		case blockAdded := <-client.onBlockAdded:
			hash := blockAdded.header.BlockHash()
			block, rawBlock, err := fetchBlock(client, hash)
			if err != nil {
				log.Warnf("Could not fetch block %s: %s", hash, err)
				continue
			}
			err = insertBlock(client, db, block, *rawBlock)
			if err != nil {
				log.Warnf("Could not insert block %s: %s", hash, err)
				continue
			}
			log.Infof("Added block %s", hash)
		case chainChanged := <-client.onChainChanged:
			removedHashes := make([]string, len(chainChanged.removedChainBlockHashes))
			for i, hash := range chainChanged.removedChainBlockHashes {
				removedHashes[i] = hash.String()
			}
			addedBlocks := make([]btcjson.ChainBlock, len(chainChanged.addedChainBlocks))
			for i, addedBlock := range chainChanged.addedChainBlocks {
				acceptedBlocks := make([]btcjson.AcceptedBlock, len(addedBlock.AcceptedBlocks))
				for j, acceptedBlock := range addedBlock.AcceptedBlocks {
					acceptedTxIDs := make([]string, len(acceptedBlock.AcceptedTxIDs))
					for k, acceptedTxID := range acceptedBlock.AcceptedTxIDs {
						acceptedTxIDs[k] = acceptedTxID.String()
					}
					acceptedBlocks[j] = btcjson.AcceptedBlock{
						Hash:          acceptedBlock.Hash.String(),
						AcceptedTxIDs: acceptedTxIDs,
					}
				}
				addedBlocks[i] = btcjson.ChainBlock{
					Hash:           addedBlock.Hash.String(),
					AcceptedBlocks: acceptedBlocks,
				}
			}
			err := updateSelectedParentChain(db, removedHashes, addedBlocks)
			if err != nil {
				log.Warnf("Could not update selected parent chain: %s", err)
			}
		case <-doneChan:
			log.Infof("blockLoop stopped")
			break loop
		}
	}
	return nil
}

func findMostRecentBlockHash(db *gorm.DB) *string {
	var block models.Block
	db.Order("blue_score DESC").First(&block)

	if block.ID == 0 {
		return nil
	}
	return &block.BlockHash
}

func SyncBlocks(client *apiServerClient, db *gorm.DB, startHash *string) error {
	var blocks []string
	var rawBlocks []btcjson.GetBlockVerboseResult
	for {
		BlocksResult, err := client.GetBlocks(true, false, startHash)
		if err != nil {
			return err
		}
		if len(BlocksResult.Hashes) == 0 {
			break
		}

		RawBlocksResult, err := client.GetBlocks(true, true, startHash)
		if err != nil {
			return err
		}

		startHash = &BlocksResult.Hashes[len(BlocksResult.Hashes)-1]
		blocks = append(blocks, BlocksResult.Blocks...)
		rawBlocks = append(rawBlocks, RawBlocksResult.RawBlocks...)
	}

	return insertBlocks(client, db, blocks, rawBlocks)
}

func SyncSelectedParentChain(client *apiServerClient, db *gorm.DB, startHash *string) error {
	for {
		chainFromBlockResult, err := client.GetChainFromBlock(false, startHash)
		if err != nil {
			return err
		}
		if len(chainFromBlockResult.AddedChainBlocks) == 0 {
			break
		}

		startHash = &chainFromBlockResult.AddedChainBlocks[len(chainFromBlockResult.AddedChainBlocks)].Hash
		err = updateSelectedParentChain(db,
			chainFromBlockResult.RemovedChainBlockHashes, chainFromBlockResult.AddedChainBlocks)
		if err != nil {
			return err
		}
	}
	return nil
}

func fetchBlock(client *apiServerClient, blockHash *daghash.Hash) (
	block string, rawBlock *btcjson.GetBlockVerboseResult, err error) {
	msgBlock, err := client.GetBlock(blockHash, nil)
	if err != nil {
		return "", nil, err
	}
	writer := bytes.NewBuffer(make([]byte, 0, msgBlock.SerializeSize()))
	err = msgBlock.Serialize(writer)
	if err != nil {
		return "", nil, err
	}
	block = hex.EncodeToString(writer.Bytes())

	rawBlock, err = client.GetBlockVerboseTx(blockHash, nil)
	if err != nil {
		return "", nil, err
	}
	return block, rawBlock, nil
}

func insertBlocks(client *apiServerClient, db *gorm.DB, blocks []string, rawBlocks []btcjson.GetBlockVerboseResult) error {
	for i, rawBlock := range rawBlocks {
		block := blocks[i]
		err := insertBlock(client, db, block, rawBlock)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertBlock(client *apiServerClient, db *gorm.DB, block string, rawBlock btcjson.GetBlockVerboseResult) error {
	db = db.Begin()

	// Insert the block
	var dbBlock models.Block
	db.Where(&models.Block{BlockHash: rawBlock.Hash}).First(&dbBlock)
	if dbBlock.ID == 0 {
		bits, err := strconv.ParseUint(rawBlock.Bits, 16, 32)
		if err != nil {
			return err
		}
		dbBlock = models.Block{
			BlockHash:            rawBlock.Hash,
			Version:              rawBlock.Version,
			HashMerkleRoot:       rawBlock.HashMerkleRoot,
			AcceptedIDMerkleRoot: rawBlock.AcceptedIDMerkleRoot,
			UTXOCommitment:       rawBlock.UTXOCommitment,
			Timestamp:            time.Unix(rawBlock.Time, 0),
			Bits:                 uint32(bits),
			Nonce:                rawBlock.Nonce,
			BlueScore:            rawBlock.BlueScore,
			IsChainBlock:         rawBlock.IsChainBlock,
		}
		db.Create(&dbBlock)
	}

	// Insert the block parents
	for _, parentHash := range rawBlock.ParentHashes {
		var dbParent models.Block
		db.Where(&models.Block{BlockHash: parentHash}).First(&dbParent)

		var dbParentBlock models.ParentBlock
		db.Where(&models.ParentBlock{BlockID: dbBlock.ID, ParentBlockID: dbParent.ID}).First(&dbParentBlock)
		if dbParentBlock.BlockID == 0 {
			dbParentBlock = models.ParentBlock{
				BlockID:       dbBlock.ID,
				ParentBlockID: dbParent.ID,
			}
			db.Create(&dbParentBlock)
		}
	}

	// Insert the block data
	var dbRawBlock models.RawBlock
	db.Where(&models.RawBlock{BlockID: dbBlock.ID}).First(&dbRawBlock)
	if dbRawBlock.BlockID == 0 {
		blockData, err := hex.DecodeString(block)
		if err != nil {
			return err
		}
		dbRawBlock = models.RawBlock{
			BlockID:   dbBlock.ID,
			BlockData: blockData,
		}
		db.Create(&dbRawBlock)
	}

	for i, transaction := range rawBlock.RawTx {
		// Insert the subnetwork
		var dbSubnetwork models.Subnetwork
		db.Where(&models.Subnetwork{SubnetworkID: transaction.Subnetwork}).First(&dbSubnetwork)
		if dbSubnetwork.ID == 0 {
			subnetwork, err := client.GetSubnetwork(transaction.Subnetwork)
			if err != nil {
				return err
			}
			dbSubnetwork = models.Subnetwork{
				SubnetworkID: transaction.Subnetwork,
				GasLimit:     subnetwork.GasLimit,
			}
			db.Create(&dbSubnetwork)
		}

		// Insert the transaction
		var dbTransaction models.Transaction
		db.Where(&models.Transaction{TransactionID: transaction.TxID}).First(&dbTransaction)
		if dbTransaction.ID == 0 {
			var dbSubnetwork models.Subnetwork
			db.Where(&models.Subnetwork{SubnetworkID: transaction.Subnetwork}).First(&dbSubnetwork)

			payload, err := hex.DecodeString(transaction.Payload)
			if err != nil {
				return err
			}
			dbTransaction = models.Transaction{
				TransactionHash: transaction.Hash,
				TransactionID:   transaction.TxID,
				LockTime:        transaction.LockTime,
				SubnetworkID:    dbSubnetwork.ID,
				Gas:             transaction.Gas,
				PayloadHash:     transaction.PayloadHash,
				Payload:         payload,
			}
			db.Create(&dbTransaction)
		}

		// Insert the transaction block
		var dbTransactionBlock models.TransactionBlock
		db.Where(&models.TransactionBlock{TransactionID: dbTransaction.ID, BlockID: dbBlock.ID}).First(&dbTransactionBlock)
		if dbTransactionBlock.TransactionID == 0 {
			dbTransactionBlock = models.TransactionBlock{
				TransactionID: dbTransaction.ID,
				BlockID:       dbBlock.ID,
				Index:         uint32(i),
			}
			db.Create(&dbTransactionBlock)
		}

		// Insert the transaction inputs
		for _, input := range transaction.Vin {
			if input.IsCoinBase() {
				continue
			}

			var dbOutputTransaction models.Transaction
			db.Where(&models.Transaction{TransactionID: input.TxID}).First(&dbOutputTransaction)

			var dbOutputTransactionOutput models.TransactionOutput
			db.Where(&models.TransactionOutput{TransactionID: dbOutputTransaction.ID, Index: input.Vout}).First(&dbOutputTransactionOutput)

			var dbTransactionInput models.TransactionInput
			db.Where(models.TransactionInput{TransactionID: dbTransaction.ID, TransactionOutputID: dbOutputTransactionOutput.ID}).First(&dbTransactionInput)
			if dbTransactionInput.TransactionID == 0 {
				scriptSig, err := hex.DecodeString(input.ScriptSig.Hex)
				if err != nil {
					return nil
				}
				dbTransactionInput = models.TransactionInput{
					TransactionID:       dbTransaction.ID,
					TransactionOutputID: dbOutputTransactionOutput.ID,
					Index:               input.Vout,
					SignatureScript:     scriptSig,
					Sequence:            input.Sequence,
				}
			}
		}

		// Insert the transaction outputs
		for _, output := range transaction.Vout {
			var dbTransactionOutput models.TransactionOutput
			db.Where(&models.TransactionOutput{TransactionID: dbTransaction.ID, Index: output.N}).First(&dbTransactionOutput)
			if dbTransactionOutput.TransactionID == 0 {
				scriptPubKey, err := hex.DecodeString(output.ScriptPubKey.Hex)
				if err != nil {
					return err
				}
				dbTransactionOutput = models.TransactionOutput{
					TransactionID: dbTransaction.ID,
					Index:         output.N,
					Value:         output.Value,
					PkScript:      scriptPubKey,
				}
				db.Create(&dbTransactionOutput)
			}
		}
	}

	db.Commit()
	return nil
}

func updateSelectedParentChain(db *gorm.DB, removedChainHashes []string, addedChainBlocks []btcjson.ChainBlock) error {
	db = db.Begin()
	for _, removedHash := range removedChainHashes {
		log.Warnf("RRRRemoved!!! %s", removedHash)

		var dbBlock models.Block
		db.Where(&models.Block{BlockHash: removedHash}).First(&dbBlock)
		if dbBlock.ID == 0 {
			// FUCK!
		}

		var dbUTXOs []models.UTXO
		db.Where(&models.UTXO{AcceptingBlockID: dbBlock.ID}).Find(&dbUTXOs)
		for _, dbUTXO := range dbUTXOs {
			var dbTransactionOutput models.TransactionOutput
			db.Where(&models.TransactionOutput{ID: dbUTXO.TransactionOutputID}).First(&dbTransactionOutput)
			if dbTransactionOutput.ID == 0 {
				// FUCK
			}

			var dbTransaction models.Transaction
			db.Where(&models.Transaction{ID: dbTransactionOutput.TransactionID}).First(&dbTransaction)
			if dbTransaction.ID == 0 {
				// FUCK
			}

		}
	}
	for _, addedBlock := range addedChainBlocks {
		log.Warnf("AAAAdded!!! %+v", addedBlock)

		for _, acceptedBlock := range addedBlock.AcceptedBlocks {
			var dbAcceptingBlock models.Block
			db.Where(&models.Block{BlockHash: acceptedBlock.Hash}).First(dbAcceptingBlock)
			if dbAcceptingBlock.ID == 0 {
				// FUCK
			}

			for _, acceptedTxID := range acceptedBlock.AcceptedTxIDs {
				var dbTransaction models.Transaction
				db.Where(&models.Transaction{TransactionID: acceptedTxID}).First(&dbTransaction)
				if dbTransaction.ID == 0 {
					// FUCK
				}

				var dbTransactionInputs []models.TransactionInput
				db.Where(&models.TransactionInput{TransactionID: dbTransaction.ID}).Find(&dbTransactionInputs)
				for _, dbTransactionInput := range dbTransactionInputs {
					db.Delete(&models.UTXO{TransactionOutputID: dbTransactionInput.TransactionOutputID})
				}

				var dbTransactionOutputs []models.TransactionOutput
				db.Where(&models.TransactionOutput{TransactionID: dbTransaction.ID}).Find(&dbTransactionOutputs)
				for _, dbTransactionOutput := range dbTransactionOutputs {
					dbUTXO := models.UTXO{
						TransactionOutputID: dbTransactionOutput.ID,
						AcceptingBlockID:    dbAcceptingBlock.ID,
					}
					db.Create(&dbUTXO)
				}
			}
		}
	}

	db.Commit()
	return nil
}
