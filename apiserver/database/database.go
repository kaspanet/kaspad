package database

import (
	"errors"
	"fmt"
	"os"

	"github.com/daglabs/btcd/apiserver/config"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/jinzhu/gorm"

	"github.com/golang-migrate/migrate/v4"
)

// db is the API server database.
var db *gorm.DB

// DB returns a reference to the database connection
func DB() (*gorm.DB, error) {
	if db == nil {
		return nil, errors.New("Database is not connected")
	}
	return db, nil
}

// Connect connects to the database mentioned in
// config variable.
func Connect(cfg *config.Config) error {
	connectionString := buildConnectionString(cfg)
	migrator, driver, err := openMigrator(connectionString)
	if err != nil {
		return err
	}
	isCurrent, err := isCurrent(migrator, driver)
	if err != nil {
		return fmt.Errorf("Error checking whether the database is current: %s", err)
	}
	if !isCurrent {
		return fmt.Errorf("Database is not current. Please migrate" +
			" the database by running the server with --migrate flag and then run it again.")
	}

	db, err = gorm.Open("mysql", connectionString)
	if err != nil {
		return err
	}
	return nil
}

// Close closes the connection to the database
func Close() error {
	if db == nil {
		return nil
	}
	err := db.Close()
	db = nil
	return err
}

func buildConnectionString(cfg *config.Config) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True",
		cfg.DBUser, cfg.DBPassword, cfg.DBAddress, cfg.DBName)
}

// isCurrent resolves whether the database is on the latest
// version of the schema.
func isCurrent(migrator *migrate.Migrate, driver source.Driver) (bool, error) {
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

func openMigrator(connectionString string) (*migrate.Migrate, source.Driver, error) {
	driver, err := source.Open("file://migrations")
	if err != nil {
		return nil, nil, err
	}
	migrator, err := migrate.NewWithSourceInstance(
		"migrations", driver, "mysql://"+connectionString)
	if err != nil {
		return nil, nil, err
	}
	return migrator, driver, nil
}

// Migrate database to the latest version.
func Migrate(cfg *config.Config) error {
	connectionString := buildConnectionString(cfg)
	migrator, driver, err := openMigrator(connectionString)
	if err != nil {
		return err
	}
	isCurrent, err := isCurrent(migrator, driver)
	if err != nil {
		return fmt.Errorf("Error checking whether the database is current: %s", err)
	}
	if isCurrent {
		log.Infof("Database is already up-to-date")
		return nil
	}
	err = migrator.Up()
	if err != nil {
		return err
	}
	log.Infof("Migrated database to the latest version")
	return nil
}
