#!/bin/sh

# Test each package separately
# Function inlining messes with monkey patching so we disable it by passing -gcflags='-l'
# Running tests with -covermode=atomic saves us from race conditions unique to the testing environment
go list ./... | \
    xargs -n1 -I{} sh -c "go test -gcflags='-l' -timeout 60s -covermode=atomic -cover {}"

retVal=$?
if [ $retVal -ne 0 ]
then
    echo " >> tests failed"
    exit 1
else
    echo " >> tests completed successfully"
    exit 0
fi
