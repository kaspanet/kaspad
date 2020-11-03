// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"fmt"
	"strings"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/version"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/id"
	"github.com/kaspanet/kaspad/util/mstime"

	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// MaxUserAgentLen is the maximum allowed length for the user agent field in a
// version message (MsgVersion).
const MaxUserAgentLen = 256

// DefaultUserAgent for appmessage in the stack
var DefaultUserAgent = fmt.Sprintf("/kaspad:%s/", version.Version())

// MsgVersion implements the Message interface and represents a kaspa version
// message. It is used for a peer to advertise itself as soon as an outbound
// connection is made. The remote peer then uses this information along with
// its own to negotiate. The remote peer must then respond with a version
// message of its own containing the negotiated values followed by a verack
// message (MsgVerAck). This exchange must take place before any further
// communication is allowed to proceed.
type MsgVersion struct {
	baseMessage
	// Version of the protocol the node is using.
	ProtocolVersion uint32

	// The peer's network (mainnet, testnet, etc.)
	Network string

	// Bitfield which identifies the enabled services.
	Services ServiceFlag

	// Time the message was generated. This is encoded as an int64 on the appmessage.
	Timestamp mstime.Time

	// Address of the local peer.
	Address *NetAddress

	// The peer unique ID
	ID *id.ID

	// The user agent that generated messsage. This is a encoded as a varString
	// on the appmessage. This has a max length of MaxUserAgentLen.
	UserAgent string

	// The selected tip hash of the generator of the version message.
	SelectedTipHash *externalapi.DomainHash

	// Don't announce transactions to peer.
	DisableRelayTx bool

	// The subnetwork of the generator of the version message. Should be nil in full nodes
	SubnetworkID *subnetworkid.SubnetworkID
}

// HasService returns whether the specified service is supported by the peer
// that generated the message.
func (msg *MsgVersion) HasService(service ServiceFlag) bool {
	return msg.Services&service == service
}

// AddService adds service as a supported service by the peer generating the
// message.
func (msg *MsgVersion) AddService(service ServiceFlag) {
	msg.Services |= service
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgVersion) Command() MessageCommand {
	return CmdVersion
}

// NewMsgVersion returns a new kaspa version message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgVersion(addr *NetAddress, id *id.ID, network string,
	selectedTipHash *externalapi.DomainHash, subnetworkID *subnetworkid.SubnetworkID) *MsgVersion {

	// Limit the timestamp to one millisecond precision since the protocol
	// doesn't support better.
	return &MsgVersion{
		ProtocolVersion: ProtocolVersion,
		Network:         network,
		Services:        0,
		Timestamp:       mstime.Now(),
		Address:         addr,
		ID:              id,
		UserAgent:       DefaultUserAgent,
		SelectedTipHash: selectedTipHash,
		DisableRelayTx:  false,
		SubnetworkID:    subnetworkID,
	}
}

// ValidateUserAgent checks userAgent length against MaxUserAgentLen
func ValidateUserAgent(userAgent string) error {
	if len(userAgent) > MaxUserAgentLen {
		str := fmt.Sprintf("user agent too long [len %d, max %d]",
			len(userAgent), MaxUserAgentLen)
		return messageError("MsgVersion", str)
	}
	return nil
}

// AddUserAgent adds a user agent to the user agent string for the version
// message. The version string is not defined to any strict format, although
// it is recommended to use the form "major.minor.revision" e.g. "2.6.41".
func (msg *MsgVersion) AddUserAgent(name string, version string,
	comments ...string) {

	newUserAgent := fmt.Sprintf("%s:%s", name, version)
	if len(comments) != 0 {
		newUserAgent = fmt.Sprintf("%s(%s)", newUserAgent,
			strings.Join(comments, "; "))
	}
	newUserAgent = fmt.Sprintf("%s%s/", msg.UserAgent, newUserAgent)
	msg.UserAgent = newUserAgent
}
