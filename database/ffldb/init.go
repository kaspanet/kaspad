package ffldb

import (
	"fmt"

	"github.com/daglabs/btcd/database"
)

func registerDriver() {
	driver := database.Driver{
		DbType: dbType,
		Create: createDBDriver,
		Open:   openDBDriver,
	}
	if err := database.RegisterDriver(driver); err != nil {
		panic(fmt.Sprintf("Failed to regiser database driver '%s': %v",
			dbType, err))
	}
}

func init() {
	registerDriver()
}
