package externalapi

import (
	"reflect"
	"testing"
)

func InitTestBlockInfoStructsForClone() []*BlockInfo {

	tests := []*BlockInfo{
		{
			true,
			BlockStatus(0x01),
			true,
		}, {
			true,
			BlockStatus(0x02),
			false,
		}, {
			true,
			'a',
			false,
		}, {
			true,
			255,
			false,
		}, {
			true,
			0,
			false,
		},
	}
	return tests
}

func TestBlockInfo_Clone(t *testing.T) {

	blockInfos := InitTestBlockInfoStructsForClone()
	for i, blockInfo := range blockInfos {
		blockInfoClone := blockInfo.Clone()

		if !reflect.DeepEqual(blockInfo, blockInfoClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
