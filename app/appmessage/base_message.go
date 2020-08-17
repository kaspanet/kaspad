package appmessage

import "time"

type baseMessage struct {
	messageNumber uint64
	receivedAt    time.Time
}

func (b *baseMessage) MessageNumber() uint64 {
	return b.messageNumber
}

func (b *baseMessage) SetMessageNumber(messageNumber uint64) {
	b.messageNumber = messageNumber
}

func (b *baseMessage) ReceivedAt() time.Time {
	return b.receivedAt
}

func (b *baseMessage) SetReceivedAt(receivedAt time.Time) {
	b.receivedAt = receivedAt
}
