package server

import (
	"context"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/util"
)

func (s *server) changeAddress() (util.Address, error) {
	walletAddr := &walletAddress{
		index:         s.keysFile.LastUsedInternalIndex + 1,
		cosignerIndex: s.keysFile.CosignerIndex,
		keyChain:      internalKeychain,
	}
	path := s.walletAddressPath(walletAddr)
	s.keysFile.LastUsedInternalIndex++
	return libkaspawallet.Address(s.params, s.keysFile.ExtendedPublicKeys, s.keysFile.MinimumSignatures, path, s.keysFile.ECDSA)
}

func (s *server) GetReceiveAddress(_ context.Context, request *pb.GetReceiveAddressRequest) (*pb.GetReceiveAddressResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	walletAddr := &walletAddress{
		index:         s.keysFile.LastUsedExternalIndex + 1,
		cosignerIndex: s.keysFile.CosignerIndex,
		keyChain:      externalKeychain,
	}
	path := s.walletAddressPath(walletAddr)
	s.keysFile.LastUsedExternalIndex++
	address, err := libkaspawallet.Address(s.params, s.keysFile.ExtendedPublicKeys, s.keysFile.MinimumSignatures, path, s.keysFile.ECDSA)
	if err != nil {
		return nil, err
	}

	return &pb.GetReceiveAddressResponse{Address: address.String()}, nil
}

func (s *server) walletAddressString(wAddr *walletAddress) (string, error) {
	path := s.walletAddressPath(wAddr)
	addr, err := libkaspawallet.Address(s.params, s.keysFile.ExtendedPublicKeys, s.keysFile.MinimumSignatures, path, s.keysFile.ECDSA)
	if err != nil {
		return "", err
	}

	return addr.String(), nil
}

func (s *server) walletAddressPath(wAddr *walletAddress) string {
	if s.isMultisig() {
		return fmt.Sprintf("m/%d/%d/%d", wAddr.cosignerIndex, wAddr.keyChain, wAddr.index)
	}
	return fmt.Sprintf("m/%d/%d", wAddr.keyChain, wAddr.index)
}

func (s *server) isMultisig() bool {
	return len(s.keysFile.ExtendedPublicKeys) > 1
}
