#!/usr/bin/env sh

CONFIG_FILE=./config/${CONF_FILE}.yml

echo "[`date`] Starting server..."
./microservice-email -config-file ${CONFIG_FILE}
