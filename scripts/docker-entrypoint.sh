#!/usr/bin/env sh

APP_BIN=${APP_BIN:-/pb}
DB_PATH=${DB_PATH:-pb_data/data.db}

if [ -n "$FLY_APP_NAME" ]; then
  echo "From Fly.io. Hello ${FLY_APP_NAME}!"
  # Auto enable Litestream if running inside Fly.io
  if [ -n "$BUCKET_NAME" ]; then
    REPLICA_URL="s3://${BUCKET_NAME}/${DB_PATH}?endpoint=fly.storage.tigris.dev&region=auto"
  fi
fi

if [ -n "$REPLICA_URL" ]; then
  litestream restore -if-db-not-exists -if-replica-exists -o "${DB_PATH}" "${REPLICA_URL}"
  litestream replicate -exec "$APP_BIN $*" "${DB_PATH}" "${REPLICA_URL}"
  exit 0
fi

exec $APP_BIN $@
