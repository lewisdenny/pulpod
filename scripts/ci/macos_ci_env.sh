#!/usr/bin/env bash

VM_NAME=${VM_NAME:-"ubuntu-22.04-pulpod"}
VM_TEMPLATE=${VM_TEMPLATE:-"ubuntu-22.04"}

check_lima() {
    if command -v "limactl" >/dev/null 2>&1; then
        return 0
    else
        echo "Lima required but not installed"
        echo "https://lima-vm.io/docs/installation/"
        return 1
    fi
}

create_vm() {
    echo "Creating VM: $VM_NAME"
    limactl start \
        --tty=false \
        --name="$VM_NAME" \
        --containerd="none" \
        --mount="${PWD}:w" \
        template://"$VM_TEMPLATE"
    echo "VM created successfully"
}

check_vm() {
    list_output=$(limactl list 2>&1)
    if [[ "$list_output" == *"$VM_NAME"* ]]; then
        exit 0
    else
        echo "VM missing, need to create"
        exit 1
    fi
}

cleanup_vm() {
    limactl delete -f "${VM_NAME}"
}

install_task() {
    limactl shell "$VM_NAME" sudo sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/sbin
    echo "Taskfile installed"
}

install_vm_dependancies() {
    limactl shell "$VM_NAME" sudo add-apt-repository universe -y
    limactl shell "$VM_NAME" sudo apt install podman -y
}

container_go_test() {
    limactl shell "$VM_NAME" task container-go-test
}

if [ $# -ne 1 ]; then
    echo "Usage: $0 [run|check]"
    exit 1
fi

case $1 in
run)
    check_lima
    create_vm
    install_task
    install_vm_dependancies
    ;;
check)
    check_lima
    check_vm
    ;;
cleanup)
    check_lima
    cleanup_vm
    ;;
container_go_test)
    check_lima
    container_go_test
    ;;
*)
    echo "Invalid argument. Use 'run' or 'check'"
    exit 1
    ;;
esac
