package externalapi

// OutpointUTXOPair is a pair of outpoint and UTXO entry
type OutpointUTXOPair struct {
	Outpoint *DomainOutpoint
	Entry    UTXOEntry
}
