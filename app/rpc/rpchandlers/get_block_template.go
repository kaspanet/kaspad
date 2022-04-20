package rpchandlers

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/version"
)

// HandleGetBlockTemplate handles the respectively named RPC command
func HandleGetBlockTemplate(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getBlockTemplateRequest := request.(*appmessage.GetBlockTemplateRequestMessage)

	var payAddress util.Address
	var err error
	donations := make([]*appmessage.Donation, len(context.Config.Donation))
	
	if len(context.Config.Donation) > 0 {
		for i, donation := range context.Config.Donation {

			donateAddress, percent, _ := strings.Cut(donation, ",")

			percentFloat, err := strconv.ParseFloat(percent, 32)
			if err != nil {
				return nil, err
			}

			donations[i] = &appmessage.Donation{
				DonationAddress: donateAddress,
				DonationPercent: float32(percentFloat),
			}
		}
	
	}
	
	if len(donations) > 0 {
		draw := rand.Float32()*100
		cumPercent := float32(0)
		for i, donation := range donations {

			cumPercent = cumPercent + donation.DonationPercent 

			if draw <= cumPercent {
				payAddress, err = util.DecodeAddress(donation.DonationAddress, context.Config.ActiveNetParams.Prefix)
				if err != nil {
					return nil, err
				}
				break
			}

			if i == len(donations) -1 {
				payAddress, err = util.DecodeAddress(getBlockTemplateRequest.PayAddress, context.Config.ActiveNetParams.Prefix)
				if err != nil {
					return nil, err
				}
				break
			}
		}
	} else {
		payAddress, err = util.DecodeAddress(getBlockTemplateRequest.PayAddress, context.Config.ActiveNetParams.Prefix)
		if err != nil {
			errorMessage := &appmessage.GetBlockTemplateResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Could not decode address: %s", err)
			return errorMessage, nil
		}
	}

	scriptPublicKey, err := txscript.PayToAddrScript(payAddress)
	if err != nil {
		return nil, err
	}

	coinbaseData := &externalapi.DomainCoinbaseData{ScriptPublicKey: scriptPublicKey, ExtraData: []byte(version.Version() + "/" + getBlockTemplateRequest.ExtraData)}
	templateBlock, err := context.Domain.MiningManager().GetBlockTemplate(coinbaseData)
	if err != nil {
		return nil, err
	}

	if uint64(len(templateBlock.Transactions[transactionhelper.CoinbaseTransactionIndex].Payload)) > context.Config.NetParams().MaxCoinbasePayloadLength {
		errorMessage := &appmessage.GetBlockTemplateResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Coinbase payload is above max length (%d). Try to shorten the extra data.", context.Config.NetParams().MaxCoinbasePayloadLength)
		return errorMessage, nil
	}

	rpcBlock := appmessage.DomainBlockToRPCBlock(templateBlock)

	isSynced, err := context.ProtocolManager.ShouldMine()
	if err != nil {
		return nil, err
	}

	return appmessage.NewGetBlockTemplateResponseMessage(rpcBlock, isSynced, donations), nil
}
