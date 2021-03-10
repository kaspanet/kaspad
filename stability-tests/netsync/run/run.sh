#!/bin/bash

FAST_DAGS_DIR="../dags-fast"
SLOW_DAGS_DIR="../dags-slow"
mapfile -t FAST_DAGS < <( ls $FAST_DAGS_DIR)
mapfile -t SLOW_DAGS < <( ls $SLOW_DAGS_DIR)

DAGS=()

for dagArchive in "${FAST_DAGS[@]}"
do
  DAGS+=("$FAST_DAGS_DIR/$dagArchive")
done

for dagArchive in "${SLOW_DAGS[@]}"
do
  DAGS+=("$SLOW_DAGS_DIR/$dagArchive")
done

for dagArchive in "${DAGS[@]}"
do
  JSON_FILE=$FAST_DAGS_DIR/$dagArchive
  netsync --simnet --dag-file $JSON_FILE --profile=7000
  TEST_EXIT_CODE=$?
  echo "$dagArchive processed"
  if [ $TEST_EXIT_CODE -ne 0 ]; then
    echo "netsync test: FAILED"
    exit 1
  fi
done

echo "netsync test: PASSED"
exit 0
