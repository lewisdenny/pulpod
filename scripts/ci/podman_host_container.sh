#!/usr/bin/env bash

CONTAINER_NAME=${CONTAINER_NAME:-"podman_host"}
CONTAINER_IMAGE=${CONTAINER_IMAGE:-"quay.io/podman/stable"}

run_in_podman_host() {
    local command="$*"

    podman exec -it "${CONTAINER_NAME}" /bin/bash -c "$command"
    local exit_code=$?

    return $exit_code
}

check_podman() {
    if command -v "podman" >/dev/null 2>&1; then
        return 0
    else
        echo "Podman required but not installed, exiting"
        echo "https://podman.io/docs/installation"
        exit 2
    fi
}

create_container() {
    podman run -d \
        --privileged \
        -v "${PWD}":/pulpod \
        --name "${CONTAINER_NAME}" \
        "${CONTAINER_IMAGE}" \
        sleep infinity
}

check_container() {
    container_inspect=$(podman container inspect -f '{{.State.Running}}' "$CONTAINER_NAME" 2>/dev/null)
    if [[ "$container_inspect" == "true" ]]; then
        exit 0
    else
        echo "Container $CONTAINER_NAME is not running or does not exist."
        exit 1
    fi
}

cleanup_container() {
    podman rm -f "${CONTAINER_NAME}"
}

if [ $# -ne 1 ]; then
    echo "Usage: $0 [run|check]"
    exit 1
fi

case $1 in
run)
    check_podman
    create_container
    ;;
check)
    check_podman
    check_container
    ;;
cleanup)
    check_podman
    cleanup_container
    ;;
*)
    echo "Invalid argument. Use 'run' or 'check'"
    exit 1
    ;;
esac
