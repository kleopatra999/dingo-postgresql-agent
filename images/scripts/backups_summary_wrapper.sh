#!/bin/bash

set -u

indent() {
  c="s/^/backups-summary> /"
  case $(uname) in
    Darwin) sed -l "$c";; # mac/bsd sed: -l buffers on line boundaries
    *)      sed -u "$c";; # unix/gnu sed: -u unbuffered (arbitrary) chunks of data
  esac
}

BACKUPS_SUMMARY_WAITTIME=${BACKUPS_SUMMARY_WAITTIME:-60}

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PG_DATA_DIR=${DATA_VOLUME}/postgres0
patroni_env=/etc/patroni.d/.envrc

ETCD_AUTH=
if [[ "${ETCD_PASSWORD:-}X" != "X" ]]; then
  ETCD_AUTH=" -u${ETCD_USERNAME:-root}:${ETCD_PASSWORD}"
fi

function wait_for_config {
  # TODO: wait for "supervisorctl status agent" to be RUNNING
  # wait for /config/patroni.yml to ensure that all variables stored in /etc/wal-e.d/env files
  wait_message="WARN: Waiting until ${patroni_env} is created..."
  while [[ ! -f ${patroni_env} ]]; do
    if [[ "${wait_message}X" != "X" ]]; then
      echo ${wait_message} >&2
    fi
    sleep 1
    wait_message="" # only show wait_message once
  done
}

function backups_summary {
  curl -s ${ETCD_URI:?required}/v2/keys/service/${PATRONI_SCOPE}/wale-backup-list \
    -X PUT -d "value=$(wal-e backup-list 2>/dev/null)" > /dev/null
  backup_lines=$(wal-e backup-list 2>/dev/null | wc -l)
  if [[ $backup_lines -ge 2 ]]; then
    if [[ "${DEBUG:-}X" != "X" ]]; then
      echo "INFO: Backup status:"
      wal-e backup-list 2>/dev/null
    fi
  else
    echo "WARNING: No backups successful yet"
  fi
}

(
  echo Waiting for configuration from agent...
  wait_for_config
  echo Configuration acquired from agent, beginning loop for base backups...

  source $patroni_env

  if [[ "${DEBUG:-}X" != "X" ]]; then
    env | sort
  fi

  while true; do
    pg_isready >/dev/null 2>&2 || continue
    in_recovery=$(psql -tqAc "select pg_is_in_recovery()")
    if [[ "${in_recovery}" == "f" ]]; then
      backups_summary
    fi
    sleep ${BACKUPS_SUMMARY_WAITTIME}
  done
) 2>&1 | indent
