# Netsync Stability Tester
This tests that the netsync is at least 5 blocks per second.

Note: the test doesn't delete kaspad's data directory and it's the user
responsibility to delete the data directories that appear in the log.

## Running
 1. `go install kaspad`.
 2. `go install ./...`.
 3. `cd run`
 4. `./run.sh`
 