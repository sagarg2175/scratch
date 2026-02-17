#!/bin/sh

echo "⏳ Waiting for Postgres to start..."

# Wait until Postgres is reachable
while ! nc -z postgres 5432; do
  sleep 1
done

echo "✅ Postgres is up. Starting API..."

# Run your Go API
/app/main
