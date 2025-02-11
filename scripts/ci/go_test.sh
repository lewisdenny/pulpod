#!/usr/bin/env bash

VM_NAME=${VM_NAME:-"ubuntu-22.04-pulpod"}
HOST_CONTAINER_NAME=${HOST_CONTAINER_NAME:-"podman_host"}
TEST_CONTAINER_NAME=${TEST_CONTAINER_NAME:-"podman_test"}
TEST_CONTAINER_IMAGE=${TEST_CONTAINER_IMAGE:-"registry.fedoraproject.org/fedora:41"}
GO_TEST_COMMAND=${GO_TEST_COMMAND:-"go test --tags=remote,exclude_graphdriver_btrfs,btrfs_noversion,exclude_graphdriver_devicemapper -coverprofile /tmp/fmtcoverage.txt ./..."}

run_in_test_container() {
    local command="$*"

    podman exec "${HOST_CONTAINER_NAME}" bash -c "podman exec \"${TEST_CONTAINER_NAME}\" /bin/bash -c \"$command\""
    local exit_code=$?

    return $exit_code
}

go_test_container() {
    run_in_test_container "cd /pulpod && ${GO_TEST_COMMAND}"
}

go_test_local() {
    eval "$GO_TEST_COMMAND"
}

if [ $# -ne 1 ]; then
    echo "Usage: $0 [local|container]"
    exit 1
fi

case $1 in
local)
    go_test_local
    ;;
container)
    go_test_container
    ;;
*)
    echo "Invalid argument. Use 'local' or 'container'"
    exit 1
    ;;
esac
