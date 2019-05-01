#!/bin/sh

set -e

export COVERAGE_PATH="./coverage.txt"
export COVERAGE_TEMP_PATH="./coverage.tmp"

# Remove the old coverage file if exists
rm -f ${COVERAGE_PATH}

# Create a new coverage file
echo 'mode: atomic' > ${COVERAGE_PATH}

# Test each package (excluding vendor packages) separately
# Function inlining messes with monkey patching so we disable it by passing -gcflags='-l'
# Running tests with -covermode=atomic saves us from race conditions unique to the testing environment
# We write coverage for every package to a temporary file so that we may append it to one global coverage file
go list ./... | \
    grep -v "vendor" | \
    xargs -n1 -I{} sh -c "go test -gcflags='-l' -covermode=atomic -coverprofile=${COVERAGE_TEMP_PATH} {} && tail -n +2 ${COVERAGE_TEMP_PATH} >> ${COVERAGE_PATH}" | \
    tee /tmp/test

# Remove the temporary coverage file
rm -f ${COVERAGE_TEMP_PATH}

# Succeed only if everything is 100% covered
grep "ok .* 100.0% of statements" -v /tmp/test > /tmp/test2 || true
if [ -s /tmp/test2 ]
then
    echo " >> tests failed or not 100% coverage"
    exit 1
else
    echo " >> tests completed successfully"
    exit 0
fi
