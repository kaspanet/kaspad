package database2

type Database interface {
	Begin() Database

	Rollback() error

	Commit() error
}

func DB() (Database, error) {
	return nil, nil
}
