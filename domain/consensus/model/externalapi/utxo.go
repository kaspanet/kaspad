package externalapi

type UTXOOutpointPair struct {
	Outpoint *DomainOutpoint
	Entry    UTXOEntry
}
