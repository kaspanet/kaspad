#!/bin/bash
set -ex

FAST_DAGS_DIR="../dags-fast"
mapfile -t DAGS < <( ls $FAST_DAGS_DIR)

for dagArchive in "${DAGS[@]}"
do
  JSON_FILE=$FAST_DAGS_DIR/$dagArchive
  netsync --simnet --dag-file $JSON_FILE --profile=7000
  TEST_EXIT_CODE=$?
  echo "$dagArchive processed"
  if [ $TEST_EXIT_CODE -ne 0 ]; then
    echo "netsync (fast) test: FAILED"
    exit 1
  fi
  rm -rf /tmp/STABILITY_TEMP_DIR_*
done

JSON_FILE="../fast-pruning-ibd-test/dag-for-fast-pruning-ibd-test.json.gz"
netsync --devnet --dag-file $JSON_FILE --profile=7000 --override-dag-params-file=../fast-pruning-ibd-test/fast-pruning-ibd-test-params.json
TEST_EXIT_CODE=$?
echo "dag-for-fast-pruning-ibd-test.json processed"
if [ $TEST_EXIT_CODE -ne 0 ]; then
  echo "netsync (fast) test: FAILED"
  exit 1
fi
rm -rf /tmp/STABILITY_TEMP_DIR_*

echo "netsync (fast) test: PASSED"
exit 0

