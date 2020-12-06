package consensushashing

import (
	"io"

	"github.com/kaspanet/kaspad/domain/consensus/utils/serialization"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/pkg/errors"
)

// BlockHash returns the given block's hash
func BlockHash(block *externalapi.DomainBlock) *externalapi.DomainHash {
	return HeaderHash(block.Header)
}

// HeaderHash returns the given header's hash
func HeaderHash(header *externalapi.DomainBlockHeader) *externalapi.DomainHash {
	// Encode the header and double sha256 everything prior to the number of
	// transactions.
	writer := hashes.NewHashWriter()
	err := serializeHeader(writer, header)
	if err != nil {
		// It seems like this could only happen if the writer returned an error.
		// and this writer should never return an error (no allocations or possible failures)
		// the only non-writer error path here is unknown types in `WriteElement`
		panic(errors.Wrap(err, "this should never happen. SHA256's digest should never return an error"))
	}

	return writer.Finalize()
}

func serializeHeader(w io.Writer, header *externalapi.DomainBlockHeader) error {
	timestamp := header.TimeInMilliseconds

	numParents := len(header.ParentHashes)
	if err := serialization.WriteElements(w, header.Version, uint64(numParents)); err != nil {
		return err
	}
	for _, hash := range header.ParentHashes {
		if err := serialization.WriteElement(w, hash); err != nil {
			return err
		}
	}
	return serialization.WriteElements(w, &header.HashMerkleRoot, &header.AcceptedIDMerkleRoot, &header.UTXOCommitment, timestamp,
		header.Bits, header.Nonce)
}
