package ffldb

import (
	"fmt"

	"github.com/kaspanet/kaspad/database"
)

func registerDriver() {
	driver := database.Driver{
		DbType: dbType,
		Create: createDBDriver,
		Open:   openDBDriver,
	}
	if err := database.RegisterDriver(driver); err != nil {
		panic(fmt.Sprintf("Failed to regiser database driver '%s': %s",
			dbType, err))
	}
}

func init() {
	registerDriver()
}
