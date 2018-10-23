#!/bin/sh

set -e

export COVERAGE_PATH="./coverage.txt"
export COVERAGE_TEMP_PATH="./coverage.tmp"

rm -f ${COVERAGE_PATH}
echo 'mode: atomic' > ${COVERAGE_PATH}

go list ./... | \
    grep -v "vendor" | \
    xargs -n1 -I{} sh -c "go test -gcflags='-l' -timeout 20s -covermode=atomic -coverprofile=${COVERAGE_TEMP_PATH} {} && tail -n +2 ${COVERAGE_TEMP_PATH} >> ${COVERAGE_PATH}" | \
    tee /tmp/test

rm -f ${COVERAGE_TEMP_PATH}

grep "ok .* 100.0% of statements" -v /tmp/test > /tmp/test2 || true
if [ -s /tmp/test2 ]
then
    echo " >> tests failed or not 100% coverage"
    exit 1
else
    echo " >> tests completed successfully"
    exit 0
fi
