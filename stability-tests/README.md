# Stability-Test Tools
This package provides some higher-level tests for kaspad.  
These are tests that are beyond the scope of unit-tests, and some of them might take long time to run.

# Running
* To run only the fast running tests call `./install_and_test.sh`
* To include all tests call `SLOW=1 ./install_and_test.sh` (Note this will take many hours to finish)
* To run a single test cd `[test-name]/run` and call `./run.sh` 
