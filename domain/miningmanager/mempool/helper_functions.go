package mempool

func (mp *mempool) virtualDAAScore() (uint64, error) {
	virtualInfo, err := mp.consensus.GetVirtualInfo()
	if err != nil {
		return 0, err
	}
	return virtualInfo.DAAScore, nil
}
