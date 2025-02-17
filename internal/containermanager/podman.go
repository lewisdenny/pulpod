package containermanager

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/specgen"
	"github.com/openstack-tooling/pulpod/internal/config"
	"github.com/openstack-tooling/pulpod/logging"
)

// PodmanManager is a concrete implementation of ContainerManager for Podman.
type PodmanManager struct {
	clientCtx context.Context
}

// NewPodmanManager creates a new PodmanManager instance.
func NewPodmanManager(containerManagerConfig *config.ContainerManagerConfig) (ContainerManager, error) {
	socket := containerManagerConfig.Socket

	//NOTE: Should this be a function? if so, where?
	if strings.HasPrefix(socket, "unix://") {
		file := strings.TrimPrefix(socket, "unix://")

		if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	clientCtx, err := bindings.NewConnection(context.Background(), socket)
	if err != nil {
		logging.Log.Error(err)
		return nil, err
	}
	return PodmanManager{clientCtx: clientCtx}, nil
}

// TODO: Extract container spec to it's own generator function
func (p PodmanManager) CreateContainer(containerImage string, containerName string) (string, error) {
	logging.Log.Infof("Creating Podman container with: %s", containerImage)
	s := specgen.NewSpecGenerator(containerImage, false)
	s.Name = containerName
	s.Labels = map[string]string{
		"pulpodControlled": "true",
	}
	s.Command = []string{"sleep", "infinity"}

	createResponse, err := containers.CreateWithSpec(p.clientCtx, s, nil)
	if err != nil {
		logging.Log.Errorf("Failed to create Podman container: %s", s.Name)
		return "", err
	}
	return createResponse.ID, nil
}

func (p PodmanManager) StartContainer(containerID string) error {
	logging.Log.Infof("Starting Podman container: %s", containerID)
	if err := containers.Start(p.clientCtx, containerID, nil); err != nil {
		logging.Log.Errorf("Failed to start Podman container: %s\n", containerID)
		return err
	}
	logging.Log.Infof("Container %s started!", containerID)
	return nil
}

func (p PodmanManager) StopContainer(containerID string) error {
	logging.Log.Infof("Stopping Podman container: %s\n", containerID)
	if err := containers.Stop(p.clientCtx, containerID, nil); err != nil {
		logging.Log.Errorf("Failed to stop Podman container: %s\n", containerID)
		return err
	}
	logging.Log.Infof("Container %s stopped!", containerID)
	return nil
}

func (p PodmanManager) RemoveContainer(containerID string) error {
	logging.Log.Infof("Removing %s container\n", containerID)
	force := true
	depend := true
	_, err := containers.Remove(p.clientCtx, containerID, &containers.RemoveOptions{Force: &force, Depend: &depend})
	if err != nil {
		logging.Log.Errorf("Failed to remove Podman container: %s\n", containerID)
		return err
	}
	return nil
}

// FIX: This should not be exposed outside testing
// ReturnContext is used in testing
func (p PodmanManager) ReturnContext() context.Context {
	ctx := p.clientCtx
	return ctx
}

func (p PodmanManager) PullImage(imageName string) error {
	pullOptions := new(images.PullOptions).WithQuiet(true)
	_, err := images.Pull(p.clientCtx, imageName, pullOptions)
	if err != nil {
		logging.Log.Errorf("Failed to pull Podman image: %s\n", imageName)
		return err
	}
	return nil
}

// List only lists containers managed by Pulpod by filtering on
// pulpodControlled=true label
func (p PodmanManager) List() ([]string, error) {
	listOptions := new(containers.ListOptions).WithFilters(map[string][]string{
		"label": {"pulpodControlled=true"},
	})
	retunedContainers, err := containers.List(p.clientCtx, listOptions)
	if err != nil {
		logging.Log.Errorln("Failed to list Podman containers")
		return nil, err
	}

	containerList := []string{}
	for _, container := range retunedContainers {
		containerList = append(containerList, container.Names[0])
	}
	return containerList, nil
}
