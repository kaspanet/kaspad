package ff

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"
)

func TestFlatFileLocationSerialization(t *testing.T) {
	location := &flatFileLocation{
		fileNumber: 1,
		fileOffset: 2,
		dataLength: 3,
	}

	serializedLocation := serializeLocation(location)
	expectedSerializedLocation := []byte{1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0}
	if !bytes.Equal(serializedLocation, expectedSerializedLocation) {
		t.Fatalf("TestFlatFileLocationSerialization: serializeLocation "+
			"returned unexpected bytes. Want: %s, got: %s",
			hex.EncodeToString(expectedSerializedLocation), hex.EncodeToString(serializedLocation))
	}

	deserializedLocation, err := deserializeLocation(serializedLocation)
	if err != nil {
		t.Fatalf("TestFlatFileLocationSerialization: deserializeLocation "+
			"unexpectedly failed: %s", err)
	}
	if !reflect.DeepEqual(deserializedLocation, location) {
		t.Fatalf("TestFlatFileLocationSerialization: original "+
			"location and deserialized location aren't the same. Want: %v, "+
			"got: %v", location, deserializedLocation)
	}
}

func TestFlatFileLocationDeserializationErrors(t *testing.T) {
	expectedError := "unexpected serializedLocation length"

	tooShortSerializedLocation := []byte{0, 1, 2, 3, 4, 5}
	_, err := deserializeLocation(tooShortSerializedLocation)
	if err == nil {
		t.Fatalf("TestFlatFileLocationSerialization: deserializeLocation " +
			"unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("TestFlatFileLocationSerialization: deserializeLocation "+
			"returned unexpected error. Want: %s, got: %s", expectedError, err)
	}

	tooLongSerializedLocation := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	_, err = deserializeLocation(tooLongSerializedLocation)
	if err == nil {
		t.Fatalf("TestFlatFileLocationSerialization: deserializeLocation " +
			"unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("TestFlatFileLocationSerialization: deserializeLocation "+
			"returned unexpected error. Want: %s, got: %s", expectedError, err)
	}
}
