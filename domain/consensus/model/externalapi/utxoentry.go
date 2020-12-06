package externalapi

// UTXOEntry houses details about an individual transaction output in a utxo
// set such as whether or not it was contained in a coinbase tx, the blue
// score of the block that accepts the tx, its public key script, and how
// much it pays.
type UTXOEntry interface {
	Amount() uint64
	ScriptPublicKey() []byte // The public key script for the output.
	BlockBlueScore() uint64  // Blue score of the block accepting the tx.
	IsCoinbase() bool
}
