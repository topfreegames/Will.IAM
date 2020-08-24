#!/bin/sh
# wait-for-it.sh

set -e
  
dabase_host="$1"
oauth2_host="$2"

shift 2
cmd="$@"
  
## Installing dependency checkers
apk add postgresql-client curl make

## Checking dependency - Postgres
>&2 echo "Checking Postgres..."

until psql -h "$dabase_host" -U "postgres" -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done
  
>&2 echo "Postgres is up"

## Checking dependency - OAuth2 server
>&2 echo "Checking OAuth2 server..."

until curl -sL "http://$oauth2_host/.well-known/openid-configuration" -o /dev/null; do
  >&2 echo "OAuth2 server is unavailable - sleeping"
  sleep 1
done

>&2 echo "OAuth2 server is up"

## Running command
>&2 echo "Running command..."
exec $cmd