#!/usr/bin/env bash

CONTAINER_NAME=${CONTAINER_NAME:-"podman_host"}

run_in_podman_host() {
  local command="$*"

  echo "$command"
  podman exec "${CONTAINER_NAME}" /bin/bash -c "$command"
  local exit_code=$?

  return $exit_code
}

check_podman_socket() {
  run_in_podman_host "ls /run/podman/podman.sock" >/dev/null 2>&1

  return $?
}

ensure_podman_socket() {
  run_in_podman_host "mkdir -p /run/podman"
  run_in_podman_host 'podman --log-level=debug system service --time=0 unix:///run/podman/podman.sock > /tmp/podman.log 2>&1 & echo $! > /tmp/podman_service_process.pid'
}

if [ $# -ne 1 ]; then
  echo "Usage: $0 [run|check]"
  exit 1
fi

case $1 in
run)
  ensure_podman_socket
  ;;
check)
  check_podman_socket
  ;;
*)
  echo "Invalid argument. Use 'run' or 'check'"
  exit 1
  ;;
esac
