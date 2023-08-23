// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// DefaultServices describes the default services that are supported by
	// the server.
	// DefaultServices는 지원되는 기본 서비스를 설명합니다.
	// 서버.
	DefaultServices = SFNodeNetwork | SFNodeBloom | SFNodeCF
)

// ServiceFlag identifies services supported by a c4ex peer.
// ServiceFlag는 c4ex 피어가 지원하는 서비스를 식별합니다.
type ServiceFlag uint64

const (
	// SFNodeNetwork is a flag used to indicate a peer is a full node.
	// ServiceFlag는 c4ex 액세서리가 지원하는 서비스를 정의합니다.
	SFNodeNetwork ServiceFlag = 1 << iota

	// SFNodeGetUTXO is a flag used to indicate a peer supports the
	// getutxos and utxos commands (BIP0064).
	// SFNodeGetUTXO는 피어가 다음을 지원함을 나타내는 데 사용되는 플래그입니다.
	// getutxos 및 utxos 명령(BIP0064).
	SFNodeGetUTXO

	// SFNodeBloom is a flag used to indicate a peer supports bloom
	// filtering.
	// SFNodeBloom은 피어가 블룸을 지원함을 나타내는 데 사용되는 플래그입니다.
	// 필터링 중입니다.
	SFNodeBloom

	// SFNodeXthin is a flag used to indicate a peer supports xthin blocks.
	// SFNodeXthin은 피어가 xthin 블록을 지원함을 나타내는 데 사용되는 플래그입니다.
	SFNodeXthin

	// SFNodeBit5 is a flag used to indicate a peer supports a service
	// defined by bit 5.
	// SFNodeBit5는 피어가 서비스를 지원함을 나타내는 데 사용되는 플래그입니다.
	// 비트 5로 정의됩니다.
	SFNodeBit5

	// SFNodeCF is a flag used to indicate a peer supports committed
	// filters (CFs).
	// SFNodeCF는 피어가 커밋을 지원함을 나타내는 데 사용되는 플래그입니다.
	// 필터(CF).
	SFNodeCF
)

// Map of service flags back to their constant names for pretty printing.
// 서비스 플래그 맵은 예쁜 인쇄를 위해 상수 이름으로 다시 표시됩니다.
var sfStrings = map[ServiceFlag]string{
	SFNodeNetwork: "SFNodeNetwork",
	SFNodeGetUTXO: "SFNodeGetUTXO",
	SFNodeBloom:   "SFNodeBloom",
	SFNodeXthin:   "SFNodeXthin",
	SFNodeBit5:    "SFNodeBit5",
	SFNodeCF:      "SFNodeCF",
}

// orderedSFStrings is an ordered list of service flags from highest to
// lowest.
// OrderedSFStrings는 서비스 플래그를 가장 높은 것부터 순서대로 나열한 목록입니다.
// 최저.
var orderedSFStrings = []ServiceFlag{
	SFNodeNetwork,
	SFNodeGetUTXO,
	SFNodeBloom,
	SFNodeXthin,
	SFNodeBit5,
	SFNodeCF,
}

// String returns the ServiceFlag in human-readable form.
func (f ServiceFlag) String() string {
	// No flags are set.
	if f == 0 {
		return "0x0"
	}

	// Add individual bit flags.
	s := ""
	for _, flag := range orderedSFStrings {
		if f&flag == flag {
			s += sfStrings[flag] + "|"
			f -= flag
		}
	}

	// Add any remaining flags which aren't accounted for as hex.
	// 16진수로 간주되지 않는 나머지 플래그를 추가합니다.
	s = strings.TrimRight(s, "|")
	if f != 0 {
		s += "|0x" + strconv.FormatUint(uint64(f), 16)
	}
	s = strings.TrimLeft(s, "|")
	return s
}

// C4exNet represents which c4ex network a message belongs to.
// C4exNet은 메시지가 속한 Kaspa 네트워크를 나타냅니다.
type C4exNet uint32

// Constants used to indicate the message c4ex network. They can also be
// used to seek to the next message when a stream's state is unknown, but
// this package does not provide that functionality since it's generally a
// better idea to simply disconnect clients that are misbehaving over TCP.
// 메시지 c4ex 네트워크를 나타내는 데 사용되는 상수입니다. 그들은 또한
// 스트림 상태를 알 수 없을 때 다음 메시지를 찾는 데 사용되지만,
// 이 패키지는 일반적으로 다음과 같은 기능을 제공하므로 해당 기능을 제공하지 않습니다.
// TCP를 통해 오작동하는 클라이언트의 연결을 끊는 것이 더 나은 아이디어입니다.
const (
	// Mainnet represents the main c4ex network.
	// 메인넷은 주요 c4ex 네트워크를 나타냅니다.
	Mainnet C4exNet = 0x3ddcf71d

	// Testnet represents the test network.
	Testnet C4exNet = 0xddb8af8f

	// Simnet represents the simulation test network.
	Simnet C4exNet = 0x374dcf1c

	// Devnet represents the development test network.
	Devnet C4exNet = 0x732d87e1
)

// bnStrings is a map of c4ex networks back to their constant names for
// pretty printing.
// bnStrings는 c4ex 네트워크의 상수 이름으로 돌아가는 맵입니다.
// 예쁜 인쇄.
var bnStrings = map[C4exNet]string{
	Mainnet: "Mainnet",
	Testnet: "Testnet",
	Simnet:  "Simnet",
	Devnet:  "Devnet",
}

// String returns the C4exNet in human-readable form.
// 문자열은 사람이 읽을 수 있는 형식으로 C4exNet을 반환합니다.
func (n C4exNet) String() string {
	if s, ok := bnStrings[n]; ok {
		return s
	}

	return fmt.Sprintf("Unknown C4exNet (%d)", uint32(n))
}
