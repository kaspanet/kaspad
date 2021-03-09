# Kaspad Sanity tool
This tries to run kapad with different sets of arguments for sanity.

In order to get clean run for each command, the tool injects its own --datadir
argument so it will be able to clean it between runs, so it's forbidden to use
--datadir as part of the arguments set.

## Running
 1. `go install` kaspad and kaspadsanity.
 2. `cd run`
 3. `./run.sh`


