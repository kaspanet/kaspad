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
	location := &flatFileLocation{
		fileNumber: 1,
		fileOffset: 2,
		dataLength: 3,
	}

	serializedLocation := serializeLocation(location)

	expectedError := "unexpected serializedLocation length"
	_, err := deserializeLocation(serializedLocation[7:])
	if err == nil {
		t.Fatalf("TestFlatFileLocationSerialization: deserializeLocation " +
			"unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("TestFlatFileLocationSerialization: deserializeLocation "+
			"returned unexpected error. Want: %s, got: %s", expectedError, err)
	}
}
