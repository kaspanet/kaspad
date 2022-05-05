package externalapi

// DomainBlockTemplate contains a Block plus metadata related to its generation
type DomainBlockTemplate struct {
	Block                *DomainBlock
	CoinbaseData         *DomainCoinbaseData
	CoinbaseHasRedReward bool
	IsNearlySynced       bool
}

// Clone returns a clone of DomainBlockTemplate
func (bt *DomainBlockTemplate) Clone() *DomainBlockTemplate {
	return &DomainBlockTemplate{
		Block:                bt.Block.Clone(),
		CoinbaseData:         bt.CoinbaseData.Clone(),
		CoinbaseHasRedReward: bt.CoinbaseHasRedReward,
		IsNearlySynced:       bt.IsNearlySynced,
	}
}
