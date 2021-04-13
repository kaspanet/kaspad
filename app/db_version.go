package app

import (
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
)

const currentDatabaseVersion = 1

func checkDatabaseVersion(dbPath string) error {
	dbVersionFileName := path.Join(dbPath, "version")
	versionBytes, err := os.ReadFile(dbVersionFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return createDatabaseVersionFile(dbVersionFileName)
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

func createDatabaseVersionFile(dbVersionFileName string) error {
	versionFile, err := os.Create(dbVersionFileName)
	if err != nil {
		return nil
	}
	defer versionFile.Close()

	versionString := strconv.Itoa(currentDatabaseVersion)
	_, err = versionFile.Write([]byte(versionString))
	return err
}
