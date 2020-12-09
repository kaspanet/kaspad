package consensusstatemanager

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

func (csm *consensusStateManager) checkFinalityViolation(blockHash *externalapi.DomainHash) error {

	log.Tracef("checkFinalityViolation start for block %s", blockHash)
	defer log.Tracef("checkFinalityViolation end for block %s", blockHash)

	isViolatingFinality, err := csm.finalityManager.IsViolatingFinality(blockHash)
	if err != nil {
		return err
	}

	if isViolatingFinality {
		csm.blockStatusStore.Stage(blockHash, externalapi.StatusUTXOPendingVerification)
		log.Warnf("Finality Violation Detected! Block %s violates finality!", blockHash)
		return nil
	}
	log.Tracef("Block %s does not violate finality", blockHash)

	return nil
}
