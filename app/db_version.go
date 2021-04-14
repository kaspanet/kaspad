package app

import (
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
)

const currentDatabaseVersion = 1

func checkDatabaseVersion(dbPath string) (doesVersionFileExist bool, err error) {
	dbVersionFileName := versionFilePath(dbPath)
	versionBytes, err := os.ReadFile(dbVersionFileName)
	if err != nil {
		if os.IsNotExist(err) { // If version file doesn't exist, we assume that the database is new
			return false, nil
		}
		return false, err
	}

	databaseVersion, err := strconv.Atoi(string(versionBytes))
	if err != nil {
		return true, err
	}

	if databaseVersion != currentDatabaseVersion {
		// TODO: Once there's more then one database version, it might make sense to add upgrade logic at this point
		return true, errors.Errorf("Invalid database version %d. Expected version: %d", databaseVersion, currentDatabaseVersion)
	}

	return true, nil
}

func createDatabaseVersionFile(dbPath string) error {
	dbVersionFileName := versionFilePath(dbPath)

	versionFile, err := os.Create(dbVersionFileName)
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
