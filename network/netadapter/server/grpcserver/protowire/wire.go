package protowire

import (
	"github.com/kaspanet/kaspad/network/appmessage"
	"github.com/pkg/errors"
)

type converter interface {
	toDomainMessage() (appmessage.Message, error)
}

// ToDomainMessage converts a KaspadMessage to its appmessage.Message representation
func (x *KaspadMessage) ToDomainMessage() (appmessage.Message, error) {
	return x.Payload.(converter).toDomainMessage()
}

// FromDomainMessage creates a KaspadMessage from a appmessage.Message
func FromDomainMessage(message appmessage.Message) (*KaspadMessage, error) {
	payload, err := toPayload(message)
	if err != nil {
		return nil, err
	}
	return &KaspadMessage{
		Payload: payload,
	}, nil
}

func toPayload(message appmessage.Message) (isKaspadMessage_Payload, error) {
	switch message := message.(type) {
	case *appmessage.MsgAddresses:
		payload := new(KaspadMessage_Addresses)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgBlock:
		payload := new(KaspadMessage_Block)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgRequestBlockLocator:
		payload := new(KaspadMessage_RequestBlockLocator)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgBlockLocator:
		payload := new(KaspadMessage_BlockLocator)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgRequestAddresses:
		payload := new(KaspadMessage_RequestAddresses)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgRequestIBDBlocks:
		payload := new(KaspadMessage_RequestIBDBlocks)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgRequestNextIBDBlocks:
		payload := new(KaspadMessage_RequestNextIBDBlocks)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgDoneIBDBlocks:
		payload := new(KaspadMessage_DoneIBDBlocks)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgRequestRelayBlocks:
		payload := new(KaspadMessage_RequestRelayBlocks)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgRequestSelectedTip:
		payload := new(KaspadMessage_RequestSelectedTip)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgRequestTransactions:
		payload := new(KaspadMessage_RequestTransactions)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgTransactionNotFound:
		payload := new(KaspadMessage_TransactionNotFound)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgIBDBlock:
		payload := new(KaspadMessage_IbdBlock)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgInvRelayBlock:
		payload := new(KaspadMessage_InvRelayBlock)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgInvTransaction:
		payload := new(KaspadMessage_InvTransactions)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgPing:
		payload := new(KaspadMessage_Ping)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgPong:
		payload := new(KaspadMessage_Pong)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgSelectedTip:
		payload := new(KaspadMessage_SelectedTip)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgTx:
		payload := new(KaspadMessage_Transaction)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgVerAck:
		payload := new(KaspadMessage_Verack)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *appmessage.MsgVersion:
		payload := new(KaspadMessage_Version)
		err := payload.fromDomainMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	default:
		return nil, errors.Errorf("unknown message type %T", message)
	}
}
