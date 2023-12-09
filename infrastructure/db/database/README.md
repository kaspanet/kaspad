database
========

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://choosealicense.com/licenses/isc/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/zoomy-network/zoomyd/database)

Package database provides a database for kaspad.

Overview
--------
This package provides a database layer to store and retrieve data in a simple
and efficient manner.

The current backend is ffldb, which makes use of leveldb, flat files, and strict
checksums in key areas to ensure data integrity.

Implementors of additional backends are required to implement the following interfaces:

DataAccessor
------------
This defines the common interface by which data gets accessed in a generic kaspad
database. Both the Database and the Transaction interfaces (see below) implement it.

Database
--------
This defines the interface of a database that can begin transactions and close itself.

Transaction
-----------
This defines the interface of a generic kaspad database transaction.

Note: Transactions provide data consistency over the state of the database as it was
when the transaction started. There is NO guarantee that if one puts data into the
transaction then it will be available to get within the same transaction.

Cursor
------
This iterates over database entries given some bucket.
