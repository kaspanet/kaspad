package database2

type Database interface {
	Begin() Database

	Rollback() error

	Commit() error

	Get(key string) ([]byte, error)

	Put(key string, value []byte) error
}

func DB() (Database, error) {
	return nil, nil
}
