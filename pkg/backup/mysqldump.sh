#!/bin/bash

set -eu


databases=$(mysql -u${DB_USERNAME} -p${DB_PASSWORD} -P${DB_PORT} -h ${DB_HOSTNAME} --skip-column-names -e "SELECT GROUP_CONCAT(schema_name SEPARATOR ' ') FROM information_schema.schemata WHERE schema_name NOT IN ('mysql','performance_schema','information_schema', 'PERCONA_SCHEMA');")
mysqldump -u${DB_USERNAME} -p${DB_PASSWORD} -P${DB_PORT} -h ${DB_HOSTNAME} -r$1 --quick --compress --opt --add-drop-database --single-transaction --databases ${databases}
