package bucket

import (
	"reflect"
	"testing"
)

func TestBuildBucketKey(t *testing.T) {
	tests := []struct {
		buckets     [][]byte
		expectedKey []byte
	}{
		{
			buckets:     [][]byte{[]byte("hello")},
			expectedKey: []byte("hello/"),
		},
		{
			buckets:     [][]byte{[]byte("hello"), []byte("world")},
			expectedKey: []byte("hello/world/"),
		},
	}

	for _, test := range tests {
		resultKey := buildBucketKey(test.buckets...)
		if !reflect.DeepEqual(resultKey, test.expectedKey) {
			t.Errorf("TestBuildBucketKey: got wrong key. Want: %s, got: %s",
				string(test.expectedKey), string(resultKey))
		}
	}
}
