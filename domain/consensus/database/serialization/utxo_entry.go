package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/pkg/errors"
	"math"
)

// ScriptPublicKeyToDBScriptPublicKey converts ScriptPublicKey to DBScriptPublicKey
func ScriptPublicKeyToDBScriptPublicKey(scriptPublicKey *externalapi.ScriptPublicKey) *DbScriptPublicKey {
	return &DbScriptPublicKey{Script: scriptPublicKey.Script, Version: uint32(scriptPublicKey.Version)}
}

// DBScriptPublicKeyToScriptPublicKey convert DbScriptPublicKey ro ScriptPublicKey
func DBScriptPublicKeyToScriptPublicKey(dbScriptPublicKey *DbScriptPublicKey) (*externalapi.ScriptPublicKey, error) {
	if dbScriptPublicKey.Version > math.MaxUint16 {
		return nil, errors.Errorf("The version on ScriptPublicKey is bigger then uint16.")
	}
	return &externalapi.ScriptPublicKey{Script: dbScriptPublicKey.Script, Version: uint16(dbScriptPublicKey.Version)}, nil
}

// UTXOEntryToDBUTXOEntry converts UTXOEntry to DbUtxoEntry
func UTXOEntryToDBUTXOEntry(utxoEntry externalapi.UTXOEntry) *DbUtxoEntry {
	dbScriptPublicKey := ScriptPublicKeyToDBScriptPublicKey(utxoEntry.ScriptPublicKey())
	return &DbUtxoEntry{
		Amount:          utxoEntry.Amount(),
		ScriptPublicKey: dbScriptPublicKey,
		BlockDaaScore:   utxoEntry.BlockDAAScore(),
		IsCoinbase:      utxoEntry.IsCoinbase(),
	}
}

// DBUTXOEntryToUTXOEntry convert DbUtxoEntry ro UTXOEntry
func DBUTXOEntryToUTXOEntry(dbUtxoEntry *DbUtxoEntry) (externalapi.UTXOEntry, error) {
	scriptPublicKey, err := DBScriptPublicKeyToScriptPublicKey(dbUtxoEntry.ScriptPublicKey)
	if err != nil {
		return nil, err
	}
	return utxo.NewUTXOEntry(dbUtxoEntry.Amount, scriptPublicKey, dbUtxoEntry.IsCoinbase, dbUtxoEntry.BlockDaaScore), nil
}
