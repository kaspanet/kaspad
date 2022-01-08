package appmessage

// MsgTrustedData represents a kaspa TrustedData message
type MsgTrustedData struct {
	baseMessage

	DAAWindow    []*TrustedDataDataDAABlockV4
	GHOSTDAGData []*BlockGHOSTDAGDataHashPair
}

// Command returns the protocol command string for the message
func (msg *MsgTrustedData) Command() MessageCommand {
	return CmdTrustedData
}

// NewMsgTrustedData returns a new MsgTrustedData.
func NewMsgTrustedData() *MsgTrustedData {
	return &MsgTrustedData{}
}
