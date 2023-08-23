package dagconfig

import (
	"time"

	"github.com/c4ei/kaspad/domain/consensus/utils/constants"
)

// The documentation refers to the following constants which aren't explicated in the code:
//	d - an upper bound on the round trip time of a block
//	delta - the expected fraction of time the width of the network exceeds defaultGHOSTDAGK
//
// For more information about defaultGHOSTDAGK, and its dependency on delta and defaultTargetTimePerBlock
// please refer to the PHANTOM paper: https://eprint.iacr.org/2018/104.pdf
//
// For more information about the DAA constants defaultDifficultyAdjustmentWindowSize, defaultTimestampDeviationTolerance,
// and their relation to defaultGHOSTDAGK and defaultTargetTimePerBlock see:
// https://research.kas.pa/t/handling-timestamp-manipulations/97
//
// For more information about defaultMergeSetSizeLimit, defaultFinalityDuration and their relation to pruning, see:
// https://research.kas.pa/t/a-proposal-for-finality-in-ghostdag/66/17
// https://research.kas.pa/t/some-of-the-intuition-behind-the-design-of-the-invalidation-rules-for-pruning/95
//
// 문서에서는 코드에 설명되지 않은 다음 상수를 참조합니다.
// d - 블록 왕복 시간의 상한
// 델타 - 네트워크 너비가 기본값을 초과하는 예상 시간 비율GHOSTDAGK
//
// defaultGHOSTDAGK 및 델타 및 defaultTargetTimePerBlock에 대한 종속성에 대한 자세한 내용은
// PHANTOM 논문을 참고하세요: https://eprint.iacr.org/2018/104.pdf
//
// DAA 상수에 대한 자세한 내용은 defaultDifficultyAdjustmentWindowSize, defaultTimestampDeviationTolerance,
// 그리고 defaultGHOSTDAGK 및 defaultTargetTimePerBlock과의 관계는 다음을 참조하세요.
// https://research.kas.pa/t/handling-timestamp-manipulations/97
//
// defaultMergeSetSizeLimit, defaultFinalityDuration 및 정리와의 관계에 대한 자세한 내용은 다음을 참조하세요.
// https://research.kas.pa/t/a-proposal-for-finality-in-ghostdag/66/17
// https://research.kas.pa/t/some-of-the-intuition-behind-the-design-of-the-invalidation-rules-for-pruning/95

const (
	defaultMaxCoinbasePayloadLength = 204
	// defaultMaxBlockMass is a bound on the mass of a block, larger values increase the bound d
	// on the round trip time of a block, which affects the other parameters as described below
	// defaultMaxBlockMass는 블록 질량의 경계이며, 값이 클수록 경계 d가 증가합니다.
	// 블록의 왕복 시간에 따라 아래 설명된 대로 다른 매개변수에 영향을 미칩니다.
	defaultMaxBlockMass = 500_000
	// defaultMassPerTxByte, defaultMassPerScriptPubKeyByte and defaultMassPerSigOp define the number of grams per
	// transaction byte, script pub key byte and sig op respectively.
	// These values are used when calculating a transactions mass.
	// defaultMassPerTxByte, defaultMassPerScriptPubKeyByte 및 defaultMassPerSigOp는 당 그램 수를 정의합니다.
	// 각각 트랜잭션 바이트, 스크립트 pub 키 바이트 및 sig op.
	// 거래량을 계산할 때 사용되는 값입니다.
	defaultMassPerTxByte           = 1
	defaultMassPerScriptPubKeyByte = 10
	defaultMassPerSigOp            = 1000
	// defaultMaxBlockParents is the number of blocks any block can point to.
	// Should be about d/defaultTargetTimePerBlock where d is a bound on the round trip time of a block.
	// defaultMaxBlockParents는 모든 블록이 가리킬 수 있는 블록 수입니다.
	// d/defaultTargetTimePerBlock 정도여야 합니다. 여기서 d는 블록의 왕복 시간에 대한 경계입니다.
	defaultMaxBlockParents = 10
	// defaultGHOSTDAGK is a bound on the number of blue blocks in the anticone of a blue block. Approximates the maximal
	// width of the network.
	// Formula (1) in section 4.2 of the PHANTOM paper shows how to calculate defaultGHOSTDAGK. The delta term represents a bound
	// on the expected fraction of the network life in which the width was higher than defaultGHOSTDAGK. The current value of K
	// was calculated for d = 5 seconds and delta = 0.05.
	// defaultGHOSTDAGK는 파란색 블록의 안티콘에 있는 파란색 블록 수에 대한 경계입니다. 최대 근사치
	// 네트워크의 너비.
	// PHANTOM 논문 섹션 4.2의 공식 (1)은 기본GHOSTDAGK를 계산하는 방법을 보여줍니다. 델타 항은 경계를 나타냅니다.
	// 너비가 defaultGHOSTDAGK보다 높은 네트워크 수명의 예상 비율. K의 현재 값
	// d = 5초, 델타 = 0.05로 계산되었습니다.
	defaultGHOSTDAGK = 18
	// defaultMergeSetSizeLimit is a bound on the size of the past of a block and the size of the past
	// of its selected parent. Any block which violates this bound is invalid.
	// Should be at least an order of magnitude smaller than defaultFinalityDuration/defaultTargetTimePerBlock.
	// (Higher values make pruning attacks easier by a constant, lower values make merging after a split or a spike
	// in block take longer)
	// defaultMergeSetSizeLimit은 블록의 과거 크기와 과거 크기의 경계입니다.
	// 선택한 부모의 이 경계를 위반하는 모든 블록은 유효하지 않습니다.
	// defaultFinalityDuration/defaultTargetTimePerBlock보다 최소한 한 자릿수는 작아야 합니다.
	// (값이 높을수록 상수에 의한 가지치기 공격이 더 쉬워지고, 값이 낮을수록 분할 또는 스파이크 후에 병합됩니다.
	// 블록에서는 시간이 더 오래 걸립니다)
	defaultMergeSetSizeLimit                       = defaultGHOSTDAGK * 10
	defaultSubsidyGenesisReward                    = 1 * constants.SompiPerKaspa
	defaultPreDeflationaryPhaseBaseSubsidy         = 500 * constants.SompiPerKaspa
	defaultDeflationaryPhaseBaseSubsidy            = 440 * constants.SompiPerKaspa
	defaultCoinbasePayloadScriptPublicKeyMaxLength = 150
	// defaultDifficultyAdjustmentWindowSize is the number of blocks in a block's past used to calculate its difficulty
	// target.
	// The DAA should take the median of 2640 blocks, so in order to do that we need 2641 window size.
	// defaultDifficultyAdjustmentWindowSize는 난이도를 계산하는 데 사용된 블록의 과거 블록 수입니다.
	// 표적.
	// DAA는 2640 블록의 중앙값을 취해야 하므로 이를 위해서는 2641 창 크기가 필요합니다.
	defaultDifficultyAdjustmentWindowSize = 2641
	// defaultTimestampDeviationTolerance is the allowed deviance of an inconming block's timestamp, measured in block delays.
	// A new block can't hold a timestamp lower than the median timestamp of the (defaultTimestampDeviationTolerance*2-1) blocks
	// with highest accumulated blue work in its past, such blocks are considered invalid.
	// A new block can't hold a timestamp higher than the local system time + defaultTimestampDeviationTolerance/defaultTargetTimePerBlock,
	// such blocks are not marked as invalid but are rejected.
	// defaultTimestampDeviationTolerance는 블록 지연으로 측정된 문제가 있는 블록의 타임스탬프에 허용되는 편차입니다.
	// 새 블록은 (defaultTimestampDeviationTolerance*2-1) 블록의 중간 타임스탬프보다 낮은 타임스탬프를 보유할 수 없습니다.
	// 과거에 가장 많이 누적된 파란색 작업이 있는 경우 이러한 블록은 유효하지 않은 것으로 간주됩니다.
	// 새 블록은 로컬 시스템 시간 + defaultTimestampDeviationTolerance/defaultTargetTimePerBlock보다 높은 타임스탬프를 보유할 수 없습니다.
	// 이러한 블록은 유효하지 않은 것으로 표시되지는 않지만 거부됩니다.
	defaultTimestampDeviationTolerance = 132
	// defaultFinalityDuration is an approximate lower bound of how old the finality block is. The finality block is chosen to
	// be the newest block in the selected chain whose blue score difference from the selected tip is at least
	// defaultFinalityDuration/defaultTargetTimePerBlock.
	// The pruning block is selected similarly, with the following duration:
	//	pruning block duration =
	//		2*defaultFinalityDuration/defaultTargetTimePerBlock + 4*defaultMergeSetSizeLimit*defaultGHOSTDAGK + 2*defaultGHOSTDAGK + 2
	// defaultFinalityDuration은 최종성 블록이 얼마나 오래되었는지에 대한 대략적인 하한입니다. 최종 블록은 다음과 같이 선택됩니다.
	// 선택된 팁과의 파란색 점수 차이가 최소한인 선택된 체인의 최신 블록이 됩니다.
	// defaultFinalityDuration/defaultTargetTimePerBlock.
	// 가지치기 블록은 다음과 같은 기간으로 유사하게 선택됩니다.
	// 가지치기 블록 기간 =
	// 2*defaultFinalityDuration/defaultTargetTimePerBlock + 4*defaultMergeSetSizeLimit*defaultGHOSTDAGK + 2*defaultGHOSTDAGK + 2
	defaultFinalityDuration = 24 * time.Hour
	// defaultTargetTimePerBlock represents how much time should pass on average between two consecutive block creations.
	// Should be parametrized such that the average width of the DAG is about defaultMaxBlockParents and such that most of the
	// time the width of the DAG is at most defaultGHOSTDAGK.
	// defaultTargetTimePerBlock은 두 개의 연속 블록 생성 사이에 평균적으로 경과해야 하는 시간을 나타냅니다.
	// DAG의 평균 너비가 defaultMaxBlockParents 정도가 되도록 매개변수화해야 하며 대부분의 경우
	// DAG의 너비가 최대 기본값GHOSTDAGK인 시간입니다.
	defaultTargetTimePerBlock = 1 * time.Second

	defaultPruningProofM = 1000

	// defaultDeflationaryPhaseDaaScore is the DAA score after which the pre-deflationary period
	// switches to the deflationary period. This number is calculated as follows:
	// We define a year as 365.25 days
	// Half a year in seconds = 365.25 / 2 * 24 * 60 * 60 = 15778800
	// The network was down for three days shortly after launch
	// Three days in seconds = 3 * 24 * 60 * 60 = 259200
	// defaultDeflationaryPhaseDaaScore는 디플레이션 이전 기간 이후의 DAA 점수입니다.
	// 디플레이션 기간으로 전환합니다. 이 숫자는 다음과 같이 계산됩니다.
	// 1년을 365.25일로 정의합니다.
	// 반년(초) = 365.25 / 2 * 24 * 60 * 60 = 15778800
	// 출시 직후 3일 ​​동안 네트워크가 다운되었습니다.
	// 3일(초) = 3 * 24 * 60 * 60 = 259200
	defaultDeflationaryPhaseDaaScore = 15778800 - 259200

	defaultMergeDepth = 3600
)
