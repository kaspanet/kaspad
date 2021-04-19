package app

import (
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
)

const currentDatabaseVersion = 1

func checkDatabaseVersion(dbPath string) (err error) {
	versionFileName := versionFilePath(dbPath)

	versionBytes, err := os.ReadFile(versionFileName)
	if err != nil {
		if os.IsNotExist(err) { // If version file doesn't exist, we assume that the database is new
			return createDatabaseVersionFile(dbPath, versionFileName)
		}
		return err
	}

	databaseVersion, err := strconv.Atoi(string(versionBytes))
	if err != nil {
		return err
	}

	if databaseVersion != currentDatabaseVersion {
		// TODO: Once there's more then one database version, it might make sense to add upgrade logic at this point
		return errors.Errorf("Invalid database version %d. Expected version: %d", databaseVersion, currentDatabaseVersion)
	}

	return nil
}

func createDatabaseVersionFile(dbPath string, versionFileName string) error {
	err := os.MkdirAll(dbPath, 0700)
	if err != nil {
		return err
	}

	versionFile, err := os.Create(versionFileName)
	if err != nil {
		return nil
	}
	defer versionFile.Close()

	versionString := strconv.Itoa(currentDatabaseVersion)
	_, err = versionFile.Write([]byte(versionString))
	return err
}

func versionFilePath(dbPath string) string {
	dbVersionFileName := path.Join(dbPath, "version")
	return dbVersionFileName
}
