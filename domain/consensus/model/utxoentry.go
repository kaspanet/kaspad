package model

// UTXOEntry houses details about an individual transaction output in a utxo
// set such as whether or not it was contained in a coinbase tx, the blue
// score of the block that accepts the tx, its public key script, and how
// much it pays.
type UTXOEntry struct {
	amount          uint64
	scriptPublicKey []byte // The public key script for the output.
	blockBlueScore  uint64 // Blue score of the block accepting the tx.
	isCoinbase      bool
}
