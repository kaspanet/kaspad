package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
)

// UTXOEntryToDBUTXOEntry converts UTXOEntry to DbUtxoEntry
func UTXOEntryToDBUTXOEntry(utxoEntry externalapi.UTXOEntry) *DbUtxoEntry {
	return &DbUtxoEntry{
		Amount:          utxoEntry.Amount(),
		ScriptPublicKey: utxoEntry.ScriptPublicKey(),
		BlockBlueScore:  utxoEntry.BlockBlueScore(),
		IsCoinbase:      utxoEntry.IsCoinbase(),
	}
}

// DBUTXOEntryToUTXOEntry convert DbUtxoEntry ro UTXOEntry
func DBUTXOEntryToUTXOEntry(dbUtxoEntry *DbUtxoEntry) externalapi.UTXOEntry {
	return utxo.NewUTXOEntry(dbUtxoEntry.Amount, dbUtxoEntry.ScriptPublicKey, dbUtxoEntry.IsCoinbase, dbUtxoEntry.BlockBlueScore)
}
