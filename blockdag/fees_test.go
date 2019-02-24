package blockdag

import (
	"io"
	"reflect"
	"testing"
)

func TestFeeAccumulators(t *testing.T) {
	fees := []uint64{1, 2, 3, 4, 5, 6, 7, 0xffffffffffffffff}

	factory := newCompactFeeFactory()

	for _, fee := range fees {
		err := factory.add(fee)
		if err != nil {
			t.Fatalf("Error writing %d as tx fee: %s", fee, err)
		}
	}

	expectedData := compactFeeData{
		1, 0, 0, 0, 0, 0, 0, 0,
		2, 0, 0, 0, 0, 0, 0, 0,
		3, 0, 0, 0, 0, 0, 0, 0,
		4, 0, 0, 0, 0, 0, 0, 0,
		5, 0, 0, 0, 0, 0, 0, 0,
		6, 0, 0, 0, 0, 0, 0, 0,
		7, 0, 0, 0, 0, 0, 0, 0,
		255, 255, 255, 255, 255, 255, 255, 255,
	}
	actualData, err := factory.data()

	if err != nil {
		t.Fatalf("Error getting bytes from writer: %s", err)
	}
	if !reflect.DeepEqual(expectedData, actualData) {
		t.Errorf("Expected bytes: %v, but got: %v", expectedData, actualData)
	}

	iterator := actualData.iterator()

	for i, expectedFee := range fees {
		actualFee, err := iterator.next()
		if err != nil {
			t.Fatalf("Error getting fee for Tx#%d: %s", i, err)
		}

		if actualFee != expectedFee {
			t.Errorf("Tx #%d: Expected fee: %d, but got %d", i, expectedFee, actualFee)
		}
	}

	_, err = iterator.next()
	if err == nil {
		t.Fatal("No error from iterator.nextTxFee after done reading all transactions")
	}
	if err != io.EOF {
		t.Fatalf("Error from iterator.nextTxFee after done reading all transactions is not io.EOF: %s", err)
	}
}
