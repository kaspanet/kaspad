package server

import (
	"context"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/util"
)

func (s *server) changeAddress() (util.Address, error) {
	path := fmt.Sprintf("m/%d/%d/%d", s.keysFile.CosignerIndex, internalKeychain, s.keysFile.LastUsedInternalIndex+1)
	s.keysFile.LastUsedInternalIndex++
	return libkaspawallet.Address(s.params, s.keysFile.ExtendedPublicKeys, s.keysFile.MinimumSignatures, path, s.keysFile.ECDSA)
}

func (s *server) GetReceiveAddress(_ context.Context, request *pb.GetReceiveAddressRequest) (*pb.GetReceiveAddressResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	path := fmt.Sprintf("m/%d/%d/%d", s.keysFile.CosignerIndex, externalKeychain, s.keysFile.LastUsedExternalIndex+1)
	s.keysFile.LastUsedExternalIndex++
	address, err := libkaspawallet.Address(s.params, s.keysFile.ExtendedPublicKeys, s.keysFile.MinimumSignatures, path, s.keysFile.ECDSA)
	if err != nil {
		return nil, err
	}

	return &pb.GetReceiveAddressResponse{Address: address.String()}, nil
}
