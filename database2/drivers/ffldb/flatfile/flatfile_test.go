package flatfile

import "testing"

func TestFlatFilePath(t *testing.T) {
	tests := []struct {
		dbPath       string
		storeName    string
		fileNumber   uint32
		expectedPath string
	}{
		{
			dbPath:       "path",
			storeName:    "store",
			fileNumber:   0,
			expectedPath: "path/store-000000000.fdb",
		},
		{
			dbPath:       "path/to/database",
			storeName:    "blocks",
			fileNumber:   123456789,
			expectedPath: "path/to/database/blocks-123456789.fdb",
		},
	}

	for _, test := range tests {
		path := flatFilePath(test.dbPath, test.storeName, test.fileNumber)
		if path != test.expectedPath {
			t.Errorf("TestFlatFilePath: unexpected path. Want: %s, got: %s",
				test.expectedPath, path)
		}
	}
}
