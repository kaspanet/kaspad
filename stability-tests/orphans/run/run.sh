#!/bin/bash
rm -rf /tmp/c4exd-temp

c4exd --simnet --appdir=/tmp/c4exd-temp --profile=6061 &
C4exD_PID=$!

sleep 1

orphans --simnet -alocalhost:22511 -n20 --profile=7000
TEST_EXIT_CODE=$?

kill $C4exD_PID

wait $C4exD_PID
C4exD_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "C4exd exit code: $C4exD_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $C4exD_EXIT_CODE -eq 0 ]; then
  echo "orphans test: PASSED"
  exit 0
fi
echo "orphans test: FAILED"
exit 1
