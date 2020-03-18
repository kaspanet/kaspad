package bucket

import (
	"reflect"
	"testing"
)

func TestBuildBucketPath(t *testing.T) {
	tests := []struct {
		buckets      [][]byte
		expectedPath []byte
	}{
		{
			buckets:      [][]byte{[]byte("hello")},
			expectedPath: []byte("hello/"),
		},
		{
			buckets:      [][]byte{[]byte("hello"), []byte("world")},
			expectedPath: []byte("hello/world/"),
		},
	}

	for _, test := range tests {
		resultKey := BuildBucketPath(test.buckets...)
		if !reflect.DeepEqual(resultKey, test.expectedPath) {
			t.Errorf("TestBuildBucketPath: got wrong path. Want: %s, got: %s",
				string(test.expectedPath), string(resultKey))
		}
	}
}

func TestBuildKey(t *testing.T) {
	tests := []struct {
		key         []byte
		buckets     [][]byte
		expectedKey []byte
	}{
		{
			key:         []byte("test"),
			buckets:     [][]byte{[]byte("hello")},
			expectedKey: []byte("hello/test"),
		},
		{
			key:         []byte("test"),
			buckets:     [][]byte{[]byte("hello"), []byte("world")},
			expectedKey: []byte("hello/world/test"),
		},
	}

	for _, test := range tests {
		resultKey := BuildKey(test.key, test.buckets...)
		if !reflect.DeepEqual(resultKey, test.expectedKey) {
			t.Errorf("TestBuildKey: got wrong key. Want: %s, got: %s",
				string(test.expectedKey), string(resultKey))
		}
	}
}
