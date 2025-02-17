#!/usr/bin/env bash

HOST_CONTAINER_NAME=${HOST_CONTAINER_NAME:-"podman_host"}
TEST_CONTAINER_NAME=${TEST_CONTAINER_NAME:-"podman_test"}
TEST_CONTAINER_IMAGE=${TEST_CONTAINER_IMAGE:-"registry.fedoraproject.org/fedora:41"}

run_in_host_container() {
  local command="$*"

  podman exec "${HOST_CONTAINER_NAME}" bash -c "$command"

  local exit_code=$?

  return $exit_code
}

run_in_test_container() {
  local command="$*"

  podman exec "${HOST_CONTAINER_NAME}" bash -c "podman exec \"${TEST_CONTAINER_NAME}\" /bin/bash -c \"$command\""
  local exit_code=$?

  return $exit_code
}

create_test_container() {
  run_in_host_container "podman run -d \
        --security-opt label:type:container_runtime_t \
        -v /run/podman/podman.sock:/run/podman/podman.sock:Z \
        -v /pulpod:/pulpod \
        --name \"${TEST_CONTAINER_NAME}\" \
        \"${TEST_CONTAINER_IMAGE}\" \
        sleep infinity"
}

bootstrap_test_container() {
  run_in_test_container "dnf install -y \
        wget \
        tput \
        golang \
        gpgme-devel \
        podman-remote"
}

check_container_running() {
  container_inspect=$(run_in_host_container podman container inspect -f '{{.State.Running}}' "$TEST_CONTAINER_NAME" 2>/dev/null)
  if [[ "$container_inspect" == "true" ]]; then
    exit 0
  else
    echo "Container $TEST_CONTAINER_NAME is not running or does not exist."
    exit 1
  fi
}

if [ $# -ne 1 ]; then
  echo "Usage: $0 [run|check]"
  exit 1
fi

case $1 in
run)
  create_test_container
  bootstrap_test_container
  ;;
check)
  check_container_running
  ;;
*)
  echo "Invalid argument. Use 'run' or 'check'"
  exit 1
  ;;
esac
