package database

import (
	"reflect"
	"testing"
)

func TestBucketPath(t *testing.T) {
	tests := []struct {
		bucketByteSlices [][]byte
		expectedPath     []byte
	}{
		{
			bucketByteSlices: [][]byte{[]byte("hello")},
			expectedPath:     []byte("hello/"),
		},
		{
			bucketByteSlices: [][]byte{[]byte("hello"), []byte("world")},
			expectedPath:     []byte("hello/world/"),
		},
	}

	for _, test := range tests {
		// Build a result using the MakeBucket function alone
		resultKey := MakeBucket(test.bucketByteSlices...).Path()
		if !reflect.DeepEqual(resultKey, test.expectedPath) {
			t.Errorf("TestBucketPath: got wrong path using MakeBucket. "+
				"Want: %s, got: %s", string(test.expectedPath), string(resultKey))
		}

		// Build a result using sub-Bucket calls
		bucket := MakeBucket()
		for _, bucketBytes := range test.bucketByteSlices {
			bucket = bucket.Bucket(bucketBytes)
		}
		resultKey = bucket.Path()
		if !reflect.DeepEqual(resultKey, test.expectedPath) {
			t.Errorf("TestBucketPath: got wrong path using sub-Bucket "+
				"calls. Want: %s, got: %s", string(test.expectedPath), string(resultKey))
		}
	}
}

func TestBucketKey(t *testing.T) {
	tests := []struct {
		bucketByteSlices [][]byte
		key              []byte
		expectedKey      []byte
	}{
		{
			bucketByteSlices: [][]byte{[]byte("hello")},
			key:              []byte("test"),
			expectedKey:      []byte("hello/test"),
		},
		{
			bucketByteSlices: [][]byte{[]byte("hello"), []byte("world")},
			key:              []byte("test"),
			expectedKey:      []byte("hello/world/test"),
		},
	}

	for _, test := range tests {
		resultKey := MakeBucket(test.bucketByteSlices...).Key(test.key)
		if !reflect.DeepEqual(resultKey, test.expectedKey) {
			t.Errorf("TestBucketKey: got wrong key. Want: %s, got: %s",
				string(test.expectedKey), string(resultKey))
		}
	}
}
