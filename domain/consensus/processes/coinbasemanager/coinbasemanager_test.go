package coinbasemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"testing"
)

func TestCalcDeflationaryPeriodBlockSubsidy(t *testing.T) {
	const secondsPerMonth = 2629800
	const deflationaryPhaseDaaScore = secondsPerMonth * 6
	const deflationaryPhaseBaseSubsidy = 440 * constants.SompiPerKaspa
	coinbaseManagerInterface := New(
		nil,
		0,
		0,
		0,
		&externalapi.DomainHash{},
		deflationaryPhaseDaaScore,
		deflationaryPhaseBaseSubsidy,
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
			name:                 "half a year after start of deflationary phase",
			blockDaaScore:        deflationaryPhaseDaaScore * 2,
			expectedBlockSubsidy: deflationaryPhaseBaseSubsidy / 2,
		},
		{
			name:                 "a year after start of deflationary phase",
			blockDaaScore:        deflationaryPhaseDaaScore * 3,
			expectedBlockSubsidy: deflationaryPhaseBaseSubsidy / 4,
		},
		{
			name:                 "two years after start of deflationary phase",
			blockDaaScore:        deflationaryPhaseDaaScore * 5,
			expectedBlockSubsidy: deflationaryPhaseBaseSubsidy / 16,
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
