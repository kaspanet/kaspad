package protowire

import (
	"math"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
)

var errorNil = errors.New("a required field is nil")

func (x *Hash) toDomain() (*externalapi.DomainHash, error) {
	if x == nil {
		return nil, errors.Wrap(errorNil, "Hash is nil")
	}
	return externalapi.NewDomainHashFromByteSlice(x.Bytes)
}

func protoHashesToDomain(protoHashes []*Hash) ([]*externalapi.DomainHash, error) {
	domainHashes := make([]*externalapi.DomainHash, len(protoHashes))
	for i, protoHash := range protoHashes {
		var err error
		domainHashes[i], err = protoHash.toDomain()
		if err != nil {
			return nil, err
		}
	}
	return domainHashes, nil
}

func domainHashToProto(hash *externalapi.DomainHash) *Hash {
	return &Hash{
		Bytes: hash.ByteSlice(),
	}
}

func domainHashesToProto(hashes []*externalapi.DomainHash) []*Hash {
	protoHashes := make([]*Hash, len(hashes))
	for i, hash := range hashes {
		protoHashes[i] = domainHashToProto(hash)
	}
	return protoHashes
}

func (x *TransactionId) toDomain() (*externalapi.DomainTransactionID, error) {
	if x == nil {
		return nil, errors.Wrap(errorNil, "TransactionId is nil")
	}
	return transactionid.FromBytes(x.Bytes)
}

func protoTransactionIDsToDomain(protoIDs []*TransactionId) ([]*externalapi.DomainTransactionID, error) {
	txIDs := make([]*externalapi.DomainTransactionID, len(protoIDs))
	for i, protoID := range protoIDs {
		var err error
		txIDs[i], err = protoID.toDomain()
		if err != nil {
			return nil, err
		}
	}
	return txIDs, nil
}

func domainTransactionIDToProto(id *externalapi.DomainTransactionID) *TransactionId {
	return &TransactionId{
		Bytes: id.ByteSlice(),
	}
}

func wireTransactionIDsToProto(ids []*externalapi.DomainTransactionID) []*TransactionId {
	protoIDs := make([]*TransactionId, len(ids))
	for i, hash := range ids {
		protoIDs[i] = domainTransactionIDToProto(hash)
	}
	return protoIDs
}

func (x *SubnetworkId) toDomain() (*externalapi.DomainSubnetworkID, error) {
	if x == nil {
		return nil, errors.Wrap(errorNil, "SubnetworkId is nil")
	}
	return subnetworks.FromBytes(x.Bytes)
}

func domainSubnetworkIDToProto(id *externalapi.DomainSubnetworkID) *SubnetworkId {
	if id == nil {
		return nil
	}
	return &SubnetworkId{
		Bytes: id[:],
	}
}

func (x *NetAddress) toAppMessage() (*appmessage.NetAddress, error) {
	if x == nil {
		return nil, errors.Wrap(errorNil, "NetAddress is nil")
	}
	if x.Port > math.MaxUint16 {
		return nil, errors.Errorf("port number is larger than %d", math.MaxUint16)
	}
	return &appmessage.NetAddress{
		Timestamp: mstime.UnixMilliseconds(x.Timestamp),
		IP:        x.Ip,
		Port:      uint16(x.Port),
	}, nil
}

func appMessageNetAddressToProto(address *appmessage.NetAddress) *NetAddress {
	return &NetAddress{
		Timestamp: address.Timestamp.UnixMilliseconds(),
		Ip:        address.IP,
		Port:      uint32(address.Port),
	}
}

func (x *Outpoint) toAppMessage() (*appmessage.Outpoint, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "Outpoint is nil")
	}
	transactionID, err := x.TransactionId.toDomain()
	if err != nil {
		return nil, err
	}
	return &appmessage.Outpoint{
		TxID:  *transactionID,
		Index: x.Index,
	}, nil
}

func (x *UtxoEntry) toAppMessage() (*appmessage.UTXOEntry, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "UtxoEntry is nil")
	}
	scriptPublicKey, err := x.ScriptPublicKey.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.UTXOEntry{
		Amount:          x.Amount,
		ScriptPublicKey: scriptPublicKey,
		BlockDAAScore:   x.BlockDaaScore,
		IsCoinbase:      x.IsCoinbase,
	}, nil
}

func (x *ScriptPublicKey) toAppMessage() (*externalapi.ScriptPublicKey, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ScriptPublicKey is nil")
	}
	if x.Version > math.MaxUint16 {
		return nil, errors.Errorf("ScriptPublicKey version is bigger then uint16.")
	}
	return &externalapi.ScriptPublicKey{
		Script:  x.Script,
		Version: uint16(x.Version),
	}, nil
}
