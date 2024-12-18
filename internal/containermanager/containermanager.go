package containermanager

import (
	"context"

	"github.com/openstack-tooling/pulpod/internal/config"
)

type ContainerManager interface {
	CreateContainer(containerImage string) (string, error)
	StartContainer(containerID string) error
	PullImage(imageName string) error
	RemoveContainer(containerName string) error
	ReturnContext() context.Context
	List() ([]string, error)
}

func NewContainerManager(containerManagerConfig *config.ContainerManagerConfig) (ContainerManager, error) {
	if containerManagerConfig.Flavor == "podman" {
		return NewPodmanManager(containerManagerConfig)
	}
	return nil, nil
}
