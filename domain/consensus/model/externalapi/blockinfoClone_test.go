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
		},

		{
			true,
			BlockStatus(0x02),
			false,
		},
		{
			true,
			'a',
			false,
		},
		{
			true,
			255,
			false,
		},
		{
			true,
			0,
			false,
		},
	}
	return tests
}

func TestBlockInfo_Clone(t *testing.T) {

	blockinfos := InitTestBlockInfoStructsForClone()

	for i, blockinfo := range blockinfos {
		clone := blockinfo.Clone()

		if !reflect.DeepEqual(blockinfo, clone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
