#!/usr/bin/env bash
set -e

DB_TYPE="${DB_CONNECTION:-pgsql}"
DB_USER="${DB_USER:-smsagribd_sc}"
DB_PASS="${DB_PASS:-FffNcn6PBeFWtKB8}"
DB_NAME="${DB_NAME:-smsagribd_sc}"
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-5432}"

DIR="$(dirname "$0")"

if [ "$DB_TYPE" = "pgsql" ] || [ "$DB_TYPE" = "postgres" ]; then
    echo "Applying PostgreSQL database schema to $DB_NAME..."
    PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" < "$DIR/schema_pgsql.sql"
    if [ -f "$DIR/dump_pgsql.sql" ]; then
        echo "Applying initial PostgreSQL dump..."
        PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" < "$DIR/dump_pgsql.sql"
    fi
else
    echo "Applying MySQL database schema to $DB_NAME..."
    mysql -h "$DB_HOST" -P "${DB_PORT:-3306}" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$DIR/schema.sql"
    if [ -f "$DIR/seeds.sql" ]; then
        echo "Applying seed data..."
        mysql -h "$DB_HOST" -P "${DB_PORT:-3306}" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$DIR/seeds.sql"
    fi
fi

echo "Database setup completed successfully!"
