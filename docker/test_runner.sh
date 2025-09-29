#!/bin/sh

set -e

# The working directory is now /src, where go.mod is located.
echo "--- Preparing to run Go tests from $(pwd) ---"
MODULE_PATH=$(grep -m 1 "module" go.mod | awk '{print $2}')
echo "--- Running tests with coverage for module: ${MODULE_PATH} ---"

mkdir -p /reports

go test -v -race -count=1 -covermode=atomic -timeout 3m \
  -coverprofile=/reports/coverage.out \
  -coverpkg=${MODULE_PATH}/... \
  ${MODULE_PATH}/tests/...
TEST_EXIT_CODE=$?

echo "\n--- Generating Coverage Reports ---"

if [ -f /reports/coverage.out ] && [ $(wc -l < /reports/coverage.out) -gt 1 ]; then
    echo "--- Analyzing coverage in terminal ---"
    go tool cover -func=/reports/coverage.out
    echo "--- Generating HTML coverage report ---"
    go tool cover -html=/reports/coverage.out -o /reports/coverage.html
else
    echo "WARN: Coverage file '/reports/coverage.out' is empty or not found. This might indicate a problem with the tests."
fi

echo "--- Fixing permissions on coverage reports ---"

chown -R ${HOST_UID:-1000}:${HOST_GID:-1000} /reports

echo "=== Test run completed with exit code: $TEST_EXIT_CODE ==="
exit $TEST_EXIT_CODE