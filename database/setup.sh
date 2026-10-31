#!/usr/bin/env bash
set -e

DB_USER="${DB_USER:-smsagribd_sc}"
DB_PASS="${DB_PASS:-FffNcn6PBeFWtKB8}"
DB_NAME="${DB_NAME:-smsagribd_sc}"
DB_HOST="${DB_HOST:-127.0.0.1}"

DIR="$(dirname "$0")"

echo "Applying database schema to $DB_NAME..."
mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$DIR/schema.sql"

if [ -f "$DIR/seeds.sql" ]; then
    echo "Applying seed data..."
    mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$DIR/seeds.sql"
fi

echo "Database setup completed successfully!"
