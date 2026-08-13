#!/bin/sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required for migrations" >&2
  exit 1
fi

if [ -n "${DATABASE_NAME:-}" ]; then
  DATABASE_URL="${DATABASE_URL%/*}/${DATABASE_NAME}"
fi

for file in /app/migrations/*.sql; do
  echo "Applying $(basename "$file")"
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$file"
done

echo "FlashX migrations complete"
