/*
Package database2 provides a database for kaspad.

Overview

This package provides a database layer to store and retrieve data in a simple
and efficient manner.

The backend is ffldb, which makes use of leveldb, flat files, and strict
checksums in key areas to ensure data integrity.

Database

The main entry point is the Database interface. It exposes functionality for
transactional-based access and storage of key-value and flat-file data.

Transactions

The Transaction struct provides facilities for rolling back or committing changes
that took place while the transaction was active.
*/
package database2
