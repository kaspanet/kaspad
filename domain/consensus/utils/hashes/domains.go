package hashes

import (
	"crypto/sha256"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
)

const (
	transcationHashDomain    = "TransactionHash"
	transcationIDDomain      = "TransactionID"
	transcationSigningDomain = "TransactionSigningHash"
	blockDomain              = "BlockHash"
	proofOfWorkDomain        = "ProofOfWorkHash"
	merkleBranchDomain       = "MerkleBranchHash"
)

// NewTransactionHashWriter Returns a new HashWriter used for transaction hashes
func NewTransactionHashWriter() HashWriter {
	blake, err := blake2b.New256([]byte(transcationHashDomain))
	if err != nil {
		panic(errors.Wrapf(err, "this should never happen. %s is less than 64 bytes", transcationHashDomain))
	}
	return HashWriter{blake}
}

// NewTransactionIDWriter Returns a new HashWriter used for transaction IDs
func NewTransactionIDWriter() HashWriter {
	blake, err := blake2b.New256([]byte(transcationIDDomain))
	if err != nil {
		panic(errors.Wrapf(err, "this should never happen. %s is less than 64 bytes", transcationIDDomain))
	}
	return HashWriter{blake}
}

// NewTransactionSigningHashWriter Returns a new HashWriter used for signing on a transaction
func NewTransactionSigningHashWriter() HashWriter {
	blake, err := blake2b.New256([]byte(transcationSigningDomain))
	if err != nil {
		panic(errors.Wrapf(err, "this should never happen. %s is less than 64 bytes", transcationSigningDomain))
	}
	return HashWriter{blake}
}

// NewTransactionSigningHashECDSAWriter Returns a new HashWriter used for signing on a transaction with ECDSA
func NewTransactionSigningHashECDSAWriter() HashWriter {
	return HashWriter{sha256.New()}
}

// NewBlockHashWriter Returns a new HashWriter used for hashing blocks
func NewBlockHashWriter() HashWriter {
	blake, err := blake2b.New256([]byte(blockDomain))
	if err != nil {
		panic(errors.Wrapf(err, "this should never happen. %s is less than 64 bytes", blockDomain))
	}
	return HashWriter{blake}
}

// NewPoWHashWriter Returns a new HashWriter used for the PoW function
func NewPoWHashWriter() HashWriter {
	blake, err := blake2b.New256([]byte(proofOfWorkDomain))
	if err != nil {
		panic(errors.Wrapf(err, "this should never happen. %s is less than 64 bytes", proofOfWorkDomain))
	}
	return HashWriter{blake}
}

// NewMerkleBranchHashWriter Returns a new HashWriter used for a merkle tree branch
func NewMerkleBranchHashWriter() HashWriter {
	blake, err := blake2b.New256([]byte(merkleBranchDomain))
	if err != nil {
		panic(errors.Wrapf(err, "this should never happen. %s is less than 64 bytes", merkleBranchDomain))
	}
	return HashWriter{blake}
}
