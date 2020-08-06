package protowire

import (
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/pkg/errors"
)

type converter interface {
	toWireMessage() (domainmessage.Message, error)
}

// ToWireMessage converts a KaspadMessage to its domainmessage.Message representation
func (x *KaspadMessage) ToWireMessage() (domainmessage.Message, error) {
	return x.Payload.(converter).toWireMessage()
}

// FromWireMessage creates a KaspadMessage from a domainmessage.Message
func FromWireMessage(message domainmessage.Message) (*KaspadMessage, error) {
	payload, err := toPayload(message)
	if err != nil {
		return nil, err
	}
	return &KaspadMessage{
		Payload: payload,
	}, nil
}

func toPayload(message domainmessage.Message) (isKaspadMessage_Payload, error) {
	switch message := message.(type) {
	case *domainmessage.MsgAddresses:
		payload := new(KaspadMessage_Addresses)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgBlock:
		payload := new(KaspadMessage_Block)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgRequestBlockLocator:
		payload := new(KaspadMessage_RequestBlockLocator)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgBlockLocator:
		payload := new(KaspadMessage_BlockLocator)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgRequestAddresses:
		payload := new(KaspadMessage_RequestAddresses)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgRequestIBDBlocks:
		payload := new(KaspadMessage_RequestIBDBlocks)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgRequestNextIBDBlocks:
		payload := new(KaspadMessage_RequestNextIBDBlocks)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgDoneIBDBlocks:
		payload := new(KaspadMessage_DoneIBDBlocks)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgRequestRelayBlocks:
		payload := new(KaspadMessage_RequestRelayBlocks)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgRequestSelectedTip:
		payload := new(KaspadMessage_RequestSelectedTip)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgRequestTransactions:
		payload := new(KaspadMessage_RequestTransactions)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgTransactionNotFound:
		payload := new(KaspadMessage_TransactionNotFound)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgIBDBlock:
		payload := new(KaspadMessage_IbdBlock)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgInvRelayBlock:
		payload := new(KaspadMessage_InvRelayBlock)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgInvTransaction:
		payload := new(KaspadMessage_InvTransactions)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgPing:
		payload := new(KaspadMessage_Ping)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgPong:
		payload := new(KaspadMessage_Pong)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgSelectedTip:
		payload := new(KaspadMessage_SelectedTip)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgTx:
		payload := new(KaspadMessage_Transaction)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgVerAck:
		payload := new(KaspadMessage_Verack)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *domainmessage.MsgVersion:
		payload := new(KaspadMessage_Version)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	default:
		return nil, errors.Errorf("unknown message type %T", message)
	}
}
