package appmessage

// MsgTrustedData represents a kaspa TrustedData message
type MsgTrustedData struct {
	baseMessage

	DAAWindow    []*TrustedDataDAAHeader
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

// TrustedDataDAAHeader is an appmessage representation of externalapi.TrustedDataDataDAAHeader
type TrustedDataDAAHeader struct {
	Header       *MsgBlockHeader
	GHOSTDAGData *BlockGHOSTDAGData
}
