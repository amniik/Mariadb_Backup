#!/usr/bin/env bash

set -eu


backup_full() {
    echo "Taking a full backup"
    LAST_FULL_DATE=$(date +%d-%m-%Y)
    mariabackup \
        --defaults-file=/etc/mysql/conf.d/my.cnf \
        --backup \
        --stream=xbstream \
        --history=$LAST_FULL_DATE > $1
}

backup_full $1
