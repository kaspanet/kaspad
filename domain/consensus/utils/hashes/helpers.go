package hashes

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// PayloadHash returns the payload hash.
func PayloadHash(payload []byte) *externalapi.DomainHash {
	writer := NewPayloadHashWriter()
	writer.InfallibleWrite(payload)
	return writer.Finalize()
}
