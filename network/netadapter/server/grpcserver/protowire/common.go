package protowire

import (
	"math"

	"github.com/kaspanet/kaspad/network/appmessage"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/pkg/errors"
)

func (x *Hash) toWire() (*daghash.Hash, error) {
	return daghash.NewHash(x.Bytes)
}

func protoHashesToWire(protoHashes []*Hash) ([]*daghash.Hash, error) {
	hashes := make([]*daghash.Hash, len(protoHashes))
	for i, protoHash := range protoHashes {
		var err error
		hashes[i], err = protoHash.toWire()
		if err != nil {
			return nil, err
		}
	}
	return hashes, nil
}

func wireHashToProto(hash *daghash.Hash) *Hash {
	return &Hash{
		Bytes: hash.CloneBytes(),
	}
}

func wireHashesToProto(hashes []*daghash.Hash) []*Hash {
	protoHashes := make([]*Hash, len(hashes))
	for i, hash := range hashes {
		protoHashes[i] = wireHashToProto(hash)
	}
	return protoHashes
}

func (x *TransactionID) toWire() (*daghash.TxID, error) {
	return daghash.NewTxID(x.Bytes)
}

func protoTransactionIDsToWire(protoIDs []*TransactionID) ([]*daghash.TxID, error) {
	txIDs := make([]*daghash.TxID, len(protoIDs))
	for i, protoID := range protoIDs {
		var err error
		txIDs[i], err = protoID.toWire()
		if err != nil {
			return nil, err
		}
	}
	return txIDs, nil
}

func wireTransactionIDToProto(id *daghash.TxID) *TransactionID {
	return &TransactionID{
		Bytes: id.CloneBytes(),
	}
}

func wireTransactionIDsToProto(ids []*daghash.TxID) []*TransactionID {
	protoIDs := make([]*TransactionID, len(ids))
	for i, hash := range ids {
		protoIDs[i] = wireTransactionIDToProto(hash)
	}
	return protoIDs
}

func (x *SubnetworkID) toWire() (*subnetworkid.SubnetworkID, error) {
	if x == nil {
		return nil, nil
	}
	return subnetworkid.New(x.Bytes)
}

func wireSubnetworkIDToProto(id *subnetworkid.SubnetworkID) *SubnetworkID {
	if id == nil {
		return nil
	}
	return &SubnetworkID{
		Bytes: id.CloneBytes(),
	}
}

func (x *NetAddress) toWire() (*appmessage.NetAddress, error) {
	if x.Port > math.MaxUint16 {
		return nil, errors.Errorf("port number is larger than %d", math.MaxUint16)
	}
	return &appmessage.NetAddress{
		Timestamp: mstime.UnixMilliseconds(x.Timestamp),
		Services:  appmessage.ServiceFlag(x.Services),
		IP:        x.Ip,
		Port:      uint16(x.Port),
	}, nil
}

func wireNetAddressToProto(address *appmessage.NetAddress) *NetAddress {
	return &NetAddress{
		Timestamp: address.Timestamp.UnixMilliseconds(),
		Services:  uint64(address.Services),
		Ip:        address.IP,
		Port:      uint32(address.Port),
	}
}
