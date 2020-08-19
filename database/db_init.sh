#!/bin/bash

psql -c "CREATE USER sessionsuser WITH PASSWORD 'dev'"
psql -c "CREATE DATABASE sessionsdb"
psql -c "GRANT ALL PRIVILEGES ON DATABASE sessionsdb to sessionsuser"
psql -c "ALTER USER sessionsuser WITH SUPERUSER"

# PGPASSWORD=dev psql -d finddb -U finduser -f database.sql

cd base_schema
PGPASSWORD=dev psql -d sessionsdb -U sessionsuser -f db_setup.sql
cd ..
