package hashes

import (
	"github.com/c4ei/kaspad/domain/consensus/model/externalapi"
)

// ToStrings converts a slice of hashes into a slice of the corresponding strings
func ToStrings(hashes []*externalapi.DomainHash) []string {
	strings := make([]string, len(hashes))
	for i, hash := range hashes {
		strings[i] = hash.String()
	}
	return strings
}
