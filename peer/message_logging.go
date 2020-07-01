package peer

import (
	"fmt"
	"github.com/kaspanet/kaspad/logs"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/kaspanet/kaspad/wire"
	"strings"
)

const (
	// maxRejectReasonLen is the maximum length of a sanitized reject reason
	// that will be logged.
	maxRejectReasonLen = 250
)

// formatLockTime returns a transaction lock time as a human-readable string.
func formatLockTime(lockTime uint64) string {
	// The lock time field of a transaction is either a block blue score at
	// which the transaction is finalized or a timestamp depending on if the
	// value is before the lockTimeThreshold. When it is under the
	// threshold it is a block blue score.
	if lockTime < txscript.LockTimeThreshold {
		return fmt.Sprintf("blue score %d", lockTime)
	}

	return mstime.UnixMilliseconds(int64(lockTime)).String()
}

// invSummary returns an inventory message as a human-readable string.
func invSummary(invList []*wire.InvVect) string {
	// No inventory.
	invLen := len(invList)
	if invLen == 0 {
		return "empty"
	}

	// One inventory item.
	if invLen == 1 {
		iv := invList[0]
		switch iv.Type {
		case wire.InvTypeError:
			return fmt.Sprintf("error %s", iv.Hash)
		case wire.InvTypeBlock:
			return fmt.Sprintf("block %s", iv.Hash)
		case wire.InvTypeSyncBlock:
			return fmt.Sprintf("sync block %s", iv.Hash)
		case wire.InvTypeMissingAncestor:
			return fmt.Sprintf("missing ancestor %s", iv.Hash)
		case wire.InvTypeTx:
			return fmt.Sprintf("tx %s", iv.Hash)
		}

		return fmt.Sprintf("unknown (%d) %s", uint32(iv.Type), iv.Hash)
	}

	// More than one inv item.
	return fmt.Sprintf("size %d", invLen)
}

// sanitizeString strips any characters which are even remotely dangerous, such
// as html control characters, from the passed string. It also limits it to
// the passed maximum size, which can be 0 for unlimited. When the string is
// limited, it will also add "..." to the string to indicate it was truncated.
func sanitizeString(str string, maxLength uint) string {
	const safeChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY" +
		"Z01234567890 .,;_/:?@"

	// Strip any characters not in the safeChars string removed.
	str = strings.Map(func(r rune) rune {
		if strings.ContainsRune(safeChars, r) {
			return r
		}
		return -1
	}, str)

	// Limit the string to the max allowed length.
	if maxLength > 0 && uint(len(str)) > maxLength {
		str = str[:maxLength]
		str = str + "..."
	}
	return str
}

// messageSummary returns a human-readable string which summarizes a message.
// Not all messages have or need a summary. This is used for debug logging.
func messageSummary(msg wire.Message) string {
	switch msg := msg.(type) {
	case *wire.MsgVersion:
		return fmt.Sprintf("agent %s, pver %d, selected tip %s",
			msg.UserAgent, msg.ProtocolVersion, msg.SelectedTipHash)

	case *wire.MsgVerAck:
		// No summary.

	case *wire.MsgGetAddr:
		if msg.IncludeAllSubnetworks {
			return "all subnetworks and full nodes"
		}
		if msg.SubnetworkID == nil {
			return "full nodes"
		}
		return fmt.Sprintf("subnetwork ID %v", msg.SubnetworkID)

	case *wire.MsgAddr:
		return fmt.Sprintf("%d addr", len(msg.AddrList))

	case *wire.MsgPing:
		// No summary - perhaps add nonce.

	case *wire.MsgPong:
		// No summary - perhaps add nonce.

	case *wire.MsgTx:
		return fmt.Sprintf("hash %s, %d inputs, %d outputs, lock %s",
			msg.TxID(), len(msg.TxIn), len(msg.TxOut),
			formatLockTime(msg.LockTime))

	case *wire.MsgBlock:
		header := &msg.Header
		return fmt.Sprintf("hash %s, ver %d, %d tx, %s", msg.BlockHash(),
			header.Version, len(msg.Transactions), header.Timestamp)

	case *wire.MsgInv:
		return invSummary(msg.InvList)

	case *wire.MsgNotFound:
		return invSummary(msg.InvList)

	case *wire.MsgGetData:
		return invSummary(msg.InvList)

	case *wire.MsgGetBlockInvs:
		return fmt.Sprintf("low hash %s, high hash %s", msg.LowHash,
			msg.HighHash)

	case *wire.MsgGetBlockLocator:
		return fmt.Sprintf("high hash %s, low hash %s", msg.HighHash,
			msg.LowHash)

	case *wire.MsgBlockLocator:
		if len(msg.BlockLocatorHashes) > 0 {
			return fmt.Sprintf("locator first hash: %s, last hash: %s", msg.BlockLocatorHashes[0], msg.BlockLocatorHashes[len(msg.BlockLocatorHashes)-1])
		}
		return fmt.Sprintf("no locator")

	case *wire.MsgReject:
		// Ensure the variable length strings don't contain any
		// characters which are even remotely dangerous such as HTML
		// control characters, etc. Also limit them to sane length for
		// logging.
		rejCommand := sanitizeString(msg.Cmd, wire.CommandSize)
		rejReason := sanitizeString(msg.Reason, maxRejectReasonLen)
		summary := fmt.Sprintf("cmd %s, code %s, reason %s", rejCommand,
			msg.Code, rejReason)
		if rejCommand == wire.CmdBlock || rejCommand == wire.CmdTx {
			summary += fmt.Sprintf(", hash %s", msg.Hash)
		}
		return summary
	}

	// No summary for other messages.
	return ""
}

func messageLogLevel(msg wire.Message) logs.Level {
	switch msg.(type) {
	case *wire.MsgReject:
		return logs.LevelWarn
	default:
		return logs.LevelDebug
	}
}
