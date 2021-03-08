# Stability-Test Tools
This package provides some higher-level tests for kaspad.  
These are tests that are beyond the scope of unit-tests, and some of them might take long time to run

# Running
1. Add and push tag in the format of `vX.X.X-rcX` in kaspad repository
2. Wait for jenkins jobs "Kaspad build release candidate" and "Kaspaminer build release candidate" to complete successfully
3. Run `./run-all.sh`
