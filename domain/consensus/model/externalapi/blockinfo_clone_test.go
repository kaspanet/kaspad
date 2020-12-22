package externalapi

import (
	"reflect"
	"testing"
)

func initTestBlockInfoStructsForClone() []*BlockInfo {

	tests := []*BlockInfo{
		{
			true,
			BlockStatus(0x01),
			0,
		}, {
			true,
			BlockStatus(0x02),
			0,
		}, {
			true,
			1,
			1,
		}, {
			true,
			255,
			2,
		}, {
			true,
			0,
			3,
		},
	}
	return tests
}

func TestBlockInfo_Clone(t *testing.T) {

	blockInfos := initTestBlockInfoStructsForClone()
	for i, blockInfo := range blockInfos {
		blockInfoClone := blockInfo.Clone()
		if !reflect.DeepEqual(blockInfo, blockInfoClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
