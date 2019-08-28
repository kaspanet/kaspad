package main

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func connectToDB(cfg *config) (*gorm.DB, error) {
	connectionString := buildConnectionString(cfg)
	isCurrent, err := isCurrent(connectionString)
	if err != nil {
		return nil, fmt.Errorf("Error checking whether the database is current: %s", err)
	}
	if !isCurrent {
		return nil, fmt.Errorf("Database is not current")
	}

	db, err := gorm.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func buildConnectionString(cfg *config) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True",
		cfg.DBUser, cfg.DBPassword, cfg.DBAddress, cfg.DBName)
}

// isCurrent resolves whether the database is on the latest
// version of the schema.
func isCurrent(connectionString string) (bool, error) {
	driver, err := source.Open("file://migrations")
	if err != nil {
		return false, err
	}
	migrator, err := migrate.NewWithSourceInstance(
		"migrations", driver, "mysql://"+connectionString)
	if err != nil {
		return false, err
	}

	// Get the current version
	version, isDirty, err := migrator.Version()
	if err == migrate.ErrNilVersion {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if isDirty {
		return false, fmt.Errorf("Database is dirty")
	}

	// The database is current if Next returns ErrNotExist
	_, err = driver.Next(version)
	if pathErr, ok := err.(*os.PathError); ok {
		if pathErr.Err == os.ErrNotExist {
			return true, nil
		}
	}
	return false, err
}
