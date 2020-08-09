package protowire

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

type converter interface {
	toWireMessage() (wire.Message, error)
}

// ToWireMessage converts a KaspadMessage to its wire.Message representation
func (x *KaspadMessage) ToWireMessage() (wire.Message, error) {
	return x.Payload.(converter).toWireMessage()
}

// FromWireMessage creates a KaspadMessage from a wire.Message
func FromWireMessage(message wire.Message) (*KaspadMessage, error) {
	payload, err := toPayload(message)
	if err != nil {
		return nil, err
	}
	return &KaspadMessage{
		Payload: payload,
	}, nil
}

func toPayload(message wire.Message) (isKaspadMessage_Payload, error) {
	switch message := message.(type) {
	case *wire.MsgAddresses:
		payload := new(KaspadMessage_Addresses)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgBlock:
		payload := new(KaspadMessage_Block)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgRequestBlockLocator:
		payload := new(KaspadMessage_RequestBlockLocator)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgBlockLocator:
		payload := new(KaspadMessage_BlockLocator)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgRequestAddresses:
		payload := new(KaspadMessage_RequestAddresses)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgRequestIBDBlocks:
		payload := new(KaspadMessage_RequestIBDBlocks)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgRequestNextIBDBlocks:
		payload := new(KaspadMessage_RequestNextIBDBlocks)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgDoneIBDBlocks:
		payload := new(KaspadMessage_DoneIBDBlocks)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgRequestRelayBlocks:
		payload := new(KaspadMessage_RequestRelayBlocks)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgRequestSelectedTip:
		payload := new(KaspadMessage_RequestSelectedTip)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgRequestTransactions:
		payload := new(KaspadMessage_RequestTransactions)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgTransactionNotFound:
		payload := new(KaspadMessage_TransactionNotFound)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgIBDBlock:
		payload := new(KaspadMessage_IbdBlock)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgInvRelayBlock:
		payload := new(KaspadMessage_InvRelayBlock)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgInvTransaction:
		payload := new(KaspadMessage_InvTransactions)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgPing:
		payload := new(KaspadMessage_Ping)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgPong:
		payload := new(KaspadMessage_Pong)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgSelectedTip:
		payload := new(KaspadMessage_SelectedTip)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgTx:
		payload := new(KaspadMessage_Transaction)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgVerAck:
		payload := new(KaspadMessage_Verack)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgVersion:
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
