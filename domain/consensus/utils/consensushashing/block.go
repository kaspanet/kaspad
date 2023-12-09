package consensushashing

import (
	"io"

	"github.com/zoomy-network/zoomyd/domain/consensus/utils/serialization"

<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/hashes"
=======
>>>>>>> Stashed changes
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/hashes"
)

// BlockHash returns the given block's hash
func BlockHash(block *externalapi.DomainBlock) *externalapi.DomainHash {
	return HeaderHash(block.Header)
}

// HeaderHash returns the given header's hash
func HeaderHash(header externalapi.BaseBlockHeader) *externalapi.DomainHash {
	// Encode the header and hash everything prior to the number of
	// transactions.
	writer := hashes.NewBlockHashWriter()
	err := serializeHeader(writer, header)
	if err != nil {
		// It seems like this could only happen if the writer returned an error.
		// and this writer should never return an error (no allocations or possible failures)
		// the only non-writer error path here is unknown types in `WriteElement`
		panic(errors.Wrap(err, "this should never happen. Hash digest should never return an error"))
	}

	return writer.Finalize()
}

func serializeHeader(w io.Writer, header externalapi.BaseBlockHeader) error {
	timestamp := header.TimeInMilliseconds()
	blueWork := header.BlueWork().Bytes()

	numParents := len(header.Parents())
	if err := serialization.WriteElements(w, header.Version(), uint64(numParents)); err != nil {
		return err
	}
	for _, blockLevelParents := range header.Parents() {
		numBlockLevelParents := len(blockLevelParents)
		if err := serialization.WriteElements(w, uint64(numBlockLevelParents)); err != nil {
			return err
		}
		for _, hash := range blockLevelParents {
			if err := serialization.WriteElement(w, hash); err != nil {
				return err
			}
		}
	}
	return serialization.WriteElements(w, header.HashMerkleRoot(), header.AcceptedIDMerkleRoot(), header.UTXOCommitment(), timestamp,
		header.Bits(), header.Nonce(), header.DAAScore(), header.BlueScore(), blueWork, header.PruningPoint())
}
