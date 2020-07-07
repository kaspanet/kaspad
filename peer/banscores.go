package peer

// Ban scores for misbehaving nodes
const (
	BanScoreUnrequestedBlock           = 100
	BanScoreInvalidBlock               = 100
	BanScoreInvalidInvBlock            = 100
	BanScoreOrphanInvAsPartOfNetsync   = 100
	BanScoreMalformedBlueScoreInOrphan = 100

	BanScoreRequestNonExistingBlock = 10

	BanScoreUnrequestedSelectedTip = 20
	BanScoreUnrequestedTx          = 20
	BanScoreInvalidTx              = 100

	BanScoreMalformedMessage = 10

	BanScoreNonVersionFirstMessage = 1
	BanScoreDuplicateVersion       = 1
	BanScoreDuplicateVerack        = 1

	BanScoreSentTooManyAddresses         = 20
	BanScoreMsgAddrWithInvalidSubnetwork = 10

	BanScoreInvalidFeeFilter = 100
	BanScoreNoFilterLoaded   = 5

	BanScoreInvalidMsgGetBlockInvs = 10

	BanScoreInvalidMsgGetBlockLocator = 100

	BanScoreEmptyBlockLocator = 100

	BanScoreSentTxToBlocksOnly = 20

	BanScoreNodeBloomFlagViolation = 100

	BanScoreStallTimeout = 1

	BanScoreUnrequestedMessage = 100
)
