#!/bin/sh
set -e

echo "--- Preparing to run Go tests ---"
MODULE_PATH=$(grep -m 1 "module" go.mod | awk '{print $2}')
echo "--- Running tests with coverage for module: ${MODULE_PATH} ---"

# With the new in-process server approach, this script becomes trivial.
# We just need to run `go test` with the correct flags.
# The TestMain function inside our Go test file now handles server setup/teardown.
go test -v -race -count=1 -covermode=atomic -timeout 3m \
  -coverprofile=/app/coverage/coverage.out \
  -coverpkg=${MODULE_PATH}/... \
  ${MODULE_PATH}/tests/...
TEST_EXIT_CODE=$?

echo "\n--- Generating Coverage Reports ---"

if [ -f /app/coverage/coverage.out ] && [ $(wc -l < /app/coverage/coverage.out) -gt 1 ]; then
    echo "--- Analyzing coverage in terminal ---"
    go tool cover -func=/app/coverage/coverage.out
    echo "--- Generating HTML coverage report ---"
    go tool cover -html=/app/coverage/coverage.out -o /app/coverage/coverage.html
else
    echo "WARN: Coverage file is empty. This indicates a problem with instrumentation."
fi

echo "--- Fixing permissions on coverage reports ---"
chown -R ${HOST_UID:-1000}:${HOST_GID:-1000} /app/coverage

echo "=== Test run completed with exit code: $TEST_EXIT_CODE ==="
exit $TEST_EXIT_CODE