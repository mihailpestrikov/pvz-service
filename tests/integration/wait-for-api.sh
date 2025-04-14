#!/bin/sh

set -e

API_URL="${TEST_API_URL:-http://api-test:8080}/health"
MAX_RETRIES=30
RETRY_INTERVAL=2

echo "Waiting for API to be ready at $API_URL..."

for i in $(seq 1 $MAX_RETRIES); do
  if curl -s -f "$API_URL" > /dev/null 2>&1; then
    echo "API is ready!"
    break
  fi

  if [ $i -eq $MAX_RETRIES ]; then
    echo "API did not become ready in time. Exiting."
    exit 1
  fi

  echo "API not ready yet. Retrying in $RETRY_INTERVAL seconds... (Attempt $i/$MAX_RETRIES)"
  sleep $RETRY_INTERVAL
done

echo "Executing tests: $@"
exec "$@"