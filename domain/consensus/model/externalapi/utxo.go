package externalapi

// UTXOOutpointPair is a pair of outpoint and UTXO entry
type UTXOOutpointPair struct {
	Outpoint *DomainOutpoint
	Entry    UTXOEntry
}
