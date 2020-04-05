package ffldb

import (
	"github.com/kaspanet/kaspad/database"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestRepairFlatFiles(t *testing.T) {
	// Create a temp db to run tests against
	path, err := ioutil.TempDir("", "TestRepairFlatFiles")
	if err != nil {
		t.Fatalf("TestRepairFlatFiles: TempDir unexpectedly "+
			"failed: %s", err)
	}
	db, err := Open(path)
	if err != nil {
		t.Fatalf("TestRepairFlatFiles: Open unexpectedly "+
			"failed: %s", err)
	}
	isOpen := true
	defer func() {
		if isOpen {
			err := db.Close()
			if err != nil {
				t.Fatalf("TestRepairFlatFiles: Close unexpectedly "+
					"failed: %s", err)
			}
		}
	}()

	// Cast to ffldb since we're going to be messing with its internals
	ffldbInstance, ok := db.(*ffldb)
	if !ok {
		t.Fatalf("TestRepairFlatFiles: unexpectedly can't cast " +
			"db to ffldb")
	}

	// Append data to the same store
	storeName := "test"
	_, err = ffldbInstance.AppendToStore(storeName, []byte("data1"))
	if err != nil {
		t.Fatalf("TestRepairFlatFiles: AppendToStore unexpectedly "+
			"failed: %s", err)
	}

	// Grab the current location to test against later
	oldCurrentLocation, err := ffldbInstance.flatFileDB.CurrentLocation(storeName)
	if err != nil {
		t.Fatalf("TestRepairFlatFiles: CurrentStoreLocation "+
			"unexpectedly failed: %s", err)
	}

	// Append more data to the same store. We expect this to disappear later.
	location2, err := ffldbInstance.AppendToStore(storeName, []byte("data2"))
	if err != nil {
		t.Fatalf("TestRepairFlatFiles: AppendToStore unexpectedly "+
			"failed: %s", err)
	}

	// Manually update the current location to point to the first piece of data
	err = setCurrentStoreLocation(ffldbInstance, storeName, oldCurrentLocation)
	if err != nil {
		t.Fatalf("TestRepairFlatFiles: setCurrentStoreLocation "+
			"unexpectedly failed: %s", err)
	}

	// Reopen the database
	err = ffldbInstance.Close()
	if err != nil {
		t.Fatalf("TestRepairFlatFiles: Close unexpectedly "+
			"failed: %s", err)
	}
	isOpen = false
	db, err = Open(path)
	if err != nil {
		t.Fatalf("TestRepairFlatFiles: Open unexpectedly "+
			"failed: %s", err)
	}
	isOpen = true
	ffldbInstance, ok = db.(*ffldb)
	if !ok {
		t.Fatalf("TestRepairFlatFiles: unexpectedly can't cast " +
			"db to ffldb")
	}

	// Make sure that the current location rolled back as expected
	currentLocation, err := ffldbInstance.flatFileDB.CurrentLocation(storeName)
	if err != nil {
		t.Fatalf("TestRepairFlatFiles: CurrentStoreLocation "+
			"unexpectedly failed: %s", err)
	}
	if !reflect.DeepEqual(oldCurrentLocation, currentLocation) {
		t.Fatalf("TestRepairFlatFiles: currentLocation did " +
			"not roll back")
	}

	// Make sure that we can't get data that no longer exists
	_, err = ffldbInstance.RetrieveFromStore(storeName, location2)
	if err == nil {
		t.Fatalf("TestRepairFlatFiles: RetrieveFromStore " +
			"unexpectedly succeeded")
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("TestRepairFlatFiles: RetrieveFromStore "+
			"returned wrong error: %s", err)
	}
}

func TestFlatFileRollbackInTransactions(t *testing.T) {
	// Create a temp db to run tests against
	path, err := ioutil.TempDir("", "TestFlatFileRollbackInTransactions")
	if err != nil {
		t.Fatalf("TestFlatFileRollbackInTransactions: TempDir unexpectedly "+
			"failed: %s", err)
	}
	db, err := Open(path)
	if err != nil {
		t.Fatalf("TestFlatFileRollbackInTransactions: Open unexpectedly "+
			"failed: %s", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			t.Fatalf("TestFlatFileRollbackInTransactions: Close unexpectedly "+
				"failed: %s", err)
		}
	}()

	// Cast to ffldb since we're going to be messing with its internals
	ffldbInstance, ok := db.(*ffldb)
	if !ok {
		t.Fatalf("TestRepairFlatFiles: unexpectedly can't cast " +
			"db to ffldb")
	}

	storeName := "test"

	// Append data1 to the store
	_, err = ffldbInstance.AppendToStore(storeName, []byte("data1"))
	if err != nil {
		t.Fatalf("TestFlatFileRollbackInTransactions: AppendToStore unexpectedly "+
			"failed: %s", err)
	}

	// Grab the current location to test against later
	oldCurrentLocation, err := ffldbInstance.flatFileDB.CurrentLocation(storeName)
	if err != nil {
		t.Fatalf("TestFlatFileRollbackInTransactions: CurrentLocation "+
			"unexpectedly failed: %s", err)
	}

	// Open a transaction
	dbTx, err := db.Begin()
	if err != nil {
		t.Fatalf("TestFlatFileRollbackInTransactions: Begin "+
			"unexpectedly failed: %s", err)
	}
	defer func() {
		err := dbTx.RollbackUnlessClosed()
		if err != nil {
			t.Fatalf("TestFlatFileRollbackInTransactions: RollbackUnlessClosed "+
				"unexpectedly failed: %s", err)
		}
	}()

	// Append data2 to the store within the transaction
	_, err = dbTx.AppendToStore(storeName, []byte("data2"))
	if err != nil {
		t.Fatalf("TestFlatFileRollbackInTransactions: AppendToStore unexpectedly "+
			"failed: %s", err)
	}

	// Rollback the transaction
	err = dbTx.Rollback()
	if err != nil {
		t.Fatalf("TestFlatFileRollbackInTransactions: Rollback "+
			"unexpectedly failed: %s", err)
	}

	// Make sure that the current location rolled back as expected
	currentLocation, err := ffldbInstance.flatFileDB.CurrentLocation(storeName)
	if err != nil {
		t.Fatalf("TestFlatFileRollbackInTransactions: CurrentLocation "+
			"unexpectedly failed: %s", err)
	}
	if !reflect.DeepEqual(oldCurrentLocation, currentLocation) {
		t.Fatalf("TestFlatFileRollbackInTransactions: currentLocation did " +
			"not roll back")
	}
}
