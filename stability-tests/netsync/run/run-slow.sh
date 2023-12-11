#!/bin/bash
set -e

SLOW_DAGS_DIR="../dags-slow"
mapfile -t DAGS < <( ls $SLOW_DAGS_DIR)

for dagArchive in "${DAGS[@]}"
do
  JSON_FILE=$SLOW_DAGS_DIR/$dagArchive
  netsync --simnet --dag-file $JSON_FILE --profile=7000
  echo "$dagArchive processed"
  if [ $TEST_EXIT_CODE -ne 0 ]; then
    echo "netsync (slow) test: FAILED"
    exit 1
  fi
  rm -rf /tmp/STABILITY_TEMP_DIR_*
done

echo "netsync (slow) test: PASSED"
exit 0

