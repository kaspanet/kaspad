package coinbasemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"testing"
	"time"
)

func TestCalcDeflationaryPeriodBlockSubsidy(t *testing.T) {
	const secondsPerMonth = 2629800
	const secondsPerHalving = secondsPerMonth * 12
	const deflationaryPhaseDaaScore = secondsPerMonth * 12
	const deflationaryPhaseBaseSubsidy = 440 * constants.SompiPerKaspa
	coinbaseManagerInterface := New(
		nil,
		0,
		0,
		0,
		&externalapi.DomainHash{},
		deflationaryPhaseDaaScore,
		deflationaryPhaseBaseSubsidy,
		time.Second,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil)
	coinbaseManagerInstance := coinbaseManagerInterface.(*coinbaseManager)

	tests := []struct {
		name                 string
		blockDaaScore        uint64
		expectedBlockSubsidy uint64
	}{
		{
			name:                 "start of deflationary phase",
			blockDaaScore:        deflationaryPhaseDaaScore,
			expectedBlockSubsidy: deflationaryPhaseBaseSubsidy,
		},
		{
			name:                 "after one halving",
			blockDaaScore:        deflationaryPhaseDaaScore + secondsPerHalving,
			expectedBlockSubsidy: deflationaryPhaseBaseSubsidy / 2,
		},
		{
			name:                 "after two halvings",
			blockDaaScore:        deflationaryPhaseDaaScore + secondsPerHalving*2,
			expectedBlockSubsidy: deflationaryPhaseBaseSubsidy / 4,
		},
		{
			name:                 "after five halvings",
			blockDaaScore:        deflationaryPhaseDaaScore + secondsPerHalving*5,
			expectedBlockSubsidy: deflationaryPhaseBaseSubsidy / 32,
		},
		{
			name:                 "after 32 halvings",
			blockDaaScore:        deflationaryPhaseDaaScore + secondsPerHalving*32,
			expectedBlockSubsidy: deflationaryPhaseBaseSubsidy / 4294967296,
		},
		{
			name:                 "just before subsidy depleted",
			blockDaaScore:        deflationaryPhaseDaaScore + secondsPerHalving*35,
			expectedBlockSubsidy: 1,
		},
		{
			name:                 "after subsidy depleted",
			blockDaaScore:        deflationaryPhaseDaaScore + secondsPerHalving*36,
			expectedBlockSubsidy: 0,
		},
	}

	for _, test := range tests {
		blockSubsidy := coinbaseManagerInstance.calcDeflationaryPeriodBlockSubsidy(test.blockDaaScore)
		if blockSubsidy != test.expectedBlockSubsidy {
			t.Errorf("TestCalcDeflationaryPeriodBlockSubsidy: test '%s' failed. Want: %d, got: %d",
				test.name, test.expectedBlockSubsidy, blockSubsidy)
		}
	}
}
