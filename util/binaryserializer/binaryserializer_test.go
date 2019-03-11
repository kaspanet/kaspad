package binaryserializer

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestBinaryFreeList(t *testing.T) {

	expectedCapacity := 8
	expectedLength := 8

	first := Borrow()
	if cap(first) != expectedCapacity {
		t.Errorf("MsgTx.TestBinaryFreeList: Expected capacity for first %d, but got %d",
			expectedCapacity, cap(first))
	}
	if len(first) != expectedLength {
		t.Errorf("MsgTx.TestBinaryFreeList: Expected length for first %d, but got %d",
			expectedLength, len(first))
	}
	Return(first)

	// Borrow again, and check that the underlying array is re-used for second
	second := Borrow()
	if cap(second) != expectedCapacity {
		t.Errorf("TestBinaryFreeList: Expected capacity for second %d, but got %d",
			expectedCapacity, cap(second))
	}
	if len(second) != expectedLength {
		t.Errorf("TestBinaryFreeList: Expected length for second %d, but got %d",
			expectedLength, len(second))
	}

	firstArrayAddress := underlyingArrayAddress(first)
	secondArrayAddress := underlyingArrayAddress(second)

	if firstArrayAddress != secondArrayAddress {
		t.Errorf("First underlying array is at address %d and second at address %d, "+
			"which means memory was not re-used", firstArrayAddress, secondArrayAddress)
	}

	Return(second)

	// test there's no crash when channel is full because borrowed too much
	buffers := make([][]byte, maxItems+1)
	for i := 0; i < maxItems+1; i++ {
		buffers[i] = Borrow()
	}
	for i := 0; i < maxItems+1; i++ {
		Return(buffers[i])
	}
}

func underlyingArrayAddress(buf []byte) uint64 {
	return uint64((*reflect.SliceHeader)(unsafe.Pointer(&buf)).Data)
}
