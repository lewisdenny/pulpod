package main

import (
	"fmt"
	"os"

	"github.com/openstack-tooling/pulpod/internal/config"
	"github.com/openstack-tooling/pulpod/internal/containermanager"
	"github.com/openstack-tooling/pulpod/logging"
)

func main() {
	//cmd.Execute()

	// Load config
	config, err := config.Configure("config.toml")
	if err != nil {
		fmt.Println(err) // No logger yet
		os.Exit(1)
	}

	// Create logger
	err = logging.GetLogger(&config.LoggingConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	// Create a container manager, this could be podman, docker or k8s
	cm, err := containermanager.NewContainerManager(&config.ContainerManagerConfig)
	if err != nil {
		fmt.Println(err)
	}
	//
	// Let's pretend we are spawning a plugin
	//
	// We just want to pull, create, start, delete, exec and list containers to start with
	// Just need to add list container, what else is required for MVP

	// Pull a container image
	err = cm.PullImage("quay.io/libpod/alpine_nginx")
	if err != nil {
		fmt.Println(err)
	}

	// Create the container and return the ID
	containerID, err := cm.CreateContainer("quay.io/libpod/alpine_nginx", "foobar")
	if err != nil {
		fmt.Println(err)
	}

	err = cm.StartContainer(containerID)

	containerList, err := cm.List()
	logging.Log.Infof("List of running containers under Pulpod control: %s", containerList)

	err = cm.RemoveContainer("foobar")
}
