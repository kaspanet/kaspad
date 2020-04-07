package ff

import (
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
