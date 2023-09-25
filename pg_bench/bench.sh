#!/usr/bin/env bash

# PgBench is needed to generate some activity so that the monitoring system will produce some nice metrics
# for us to look into.
hosts=( "$@" )
host=$1
retry=0
success=false
until [ $retry -ge 10 ]; do
  echo "Trying on $host host"
  export PGHOST="$host"
  psql -c '\q' && success=true

  if [ "$success" = true ]; then
    echo "Postgres is available."
    break
  fi

  echo "Postgres is unavailable ($retry) - sleeping..."
  retry=$((retry+1))
  sleep 5
done

for i in "${hosts[@]}"
do
  echo "Trying $i host"
  export PGHOST="$i"
  is_replica=$(psql "user=postgres port=5432 dbname=postgres" -c "select pg_is_in_recovery()" -t)
  if [ "$is_replica" = " f" ]; then
  export PGHOST="$i"
  break
  fi
done

sleep 5
pgbench -i -s 10 postgres

# Run forever
while true; do
    sleep 10
    echo "Starting pgbench"
    pgbench -c 8 -T 3600 -s 5 -b tpcb-like postgres
done


