package externalapi

// VirtualInfo represents information about the virtual block needed by external components
type VirtualInfo struct {
	ParentHashes   []*DomainHash
	Bits           uint32
	PastMedianTime int64
	BlueScore      uint64
	DAAScore       uint64
}
