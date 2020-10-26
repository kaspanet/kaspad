package hashserialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/pkg/errors"
	"io"
)

func serializeHeader(w io.Writer, header *externalapi.DomainBlockHeader) error {
	timestamp := header.TimeInMilliseconds

	numParents := len(header.ParentHashes)
	if err := writeElements(w, header.Version, uint64(numParents)); err != nil {
		return err
	}
	for _, hash := range header.ParentHashes {
		if err := WriteElement(w, hash); err != nil {
			return err
		}
	}
	return writeElements(w, header.HashMerkleRoot, header.AcceptedIDMerkleRoot, header.UTXOCommitment, timestamp, header.Bits, header.Nonce)
}

// HeaderHash returns the given header hash
func HeaderHash(header *externalapi.DomainBlockHeader) *externalapi.DomainHash {
	// Encode the header and double sha256 everything prior to the number of
	// transactions.
	writer := hashes.NewHashWriter()
	err := serializeHeader(writer, header)
	if err != nil {
		// It seems like this could only happen if the writer returned an error.
		// and this writer should never return an error (no allocations or possible failures)
		// the only non-writer error path here is unknown types in `WriteElement`
		panic(errors.Wrap(err, "BlockHash() failed. this should never happen. SHA256's digest should never return an error"))
	}

	res := writer.Finalize()
	return &res
}
