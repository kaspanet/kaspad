reorg --dag-file ../../netsync/dags-slow/wide-dag-blocks--2^16-delay-factor--1-k--18.json.gz --profile=6061

TEST_EXIT_CODE=$?
echo "Exit code: $TEST_EXIT_CODE"


if [ $TEST_EXIT_CODE -eq 0 ]; then
  echo "reorg test: PASSED"
  exit 0
fi
echo "reorg test: FAILED"
exit 1
