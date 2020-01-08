/*
Package netsync implements a concurrency safe block syncing protocol. The
SyncManager communicates with connected peers to perform an initial block
download, keep the DAG and unconfirmed transaction pool in sync, and announce
new blocks connected to the DAG. Currently the sync manager selects a single
sync peer that it downloads all blocks from until it is up to date with the
selected tip of the sync peer.
*/
package netsync
