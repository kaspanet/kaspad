package consensusstatestore

import (
	"bytes"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/pkg/errors"
)

var tipsBucket = dbkeys.MakeBucket([]byte("tips"))
var tipsKey = tipsBucket.Key([]byte("tips"))

func (c consensusStateStore) Tips(dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	if c.stagedTips != nil {
		return c.stagedTips, nil
	}

	tipsBytes, err := dbContext.Get(tipsKey)
	if err != nil {
		return nil, err
	}

	return deserializeTips(tipsBytes)
}

func (c consensusStateStore) StageTips(tipHashes []*externalapi.DomainHash) error {
	c.stagedTips = tipHashes

	return nil
}

func (c consensusStateStore) commitTips(dbTx model.DBTransaction) error {
	tipsBytes := serializeTips(c.stagedTips)

	return dbTx.Put(tipsKey, tipsBytes)
}

func deserializeTips(tipsBytes []byte) ([]*externalapi.DomainHash, error) {
	if len(tipsBytes)%externalapi.DomainHashSize != 0 {
		return nil, errors.Errorf("serialized tips length is %d bytes, while it should be a multiple of %d",
			len(tipsBytes), externalapi.DomainHashSize)
	}

	tips := make([]*externalapi.DomainHash, 0, len(tipsBytes)/externalapi.DomainHashSize)

	for i := 0; i < len(tipsBytes); i += externalapi.DomainHashSize {
		tipBytes := tipsBytes[i : i+externalapi.DomainHashSize]
		tip, err := hashes.FromBytes(tipBytes)
		if err != nil {
			return nil, err
		}

		tips = append(tips, tip)
	}

	return tips, nil
}

func serializeTips(tips []*externalapi.DomainHash) []byte {
	tipsBytes := make([][]byte, 0, len(tips))

	for _, tip := range tips {
		tipsBytes = append(tipsBytes, tip[:])
	}

	return bytes.Join(tipsBytes, []byte{})
}
