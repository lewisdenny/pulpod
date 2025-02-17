package containermanager

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/bindings/system"
	"github.com/containers/podman/v5/pkg/specgen"
	"github.com/openstack-tooling/pulpod/internal/config"
	"github.com/openstack-tooling/pulpod/logging"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

var (
	err        error
	TestConfig *config.Config
)

func init() {
	// Load real config
	TestConfig, err = config.Configure("../../config.toml")
	if err != nil {
		fmt.Println(err) // No logger yet
		os.Exit(1)
	}

	// Create logger
	err = logging.GetLogger(&TestConfig.LoggingConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestNewPodmanManager_Integration(t *testing.T) {
	var (
		invalidConfig = config.ContainerManagerConfig{Socket: "unix:///fake/socket", Flavor: "podman"}

		// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
		versionPattern = `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.
		(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
	)

	testCases := []struct {
		name   string
		config *config.ContainerManagerConfig
		pass   bool
	}{
		{"test with real config", &TestConfig.ContainerManagerConfig, true},
		{"test with fake config", &invalidConfig, false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.pass {
				// Create instance of PodmanManager
				actual, err := NewPodmanManager(tc.config)

				// Assert there is no error and connection is not nil
				assert.NoError(t, err)
				assert.NotNil(t, actual.ReturnContext)

				// Fetch the version of the Podman server to test connection
				// and unsure no error is returned, note this is using a function
				// outside our project to verify.
				version, err := system.Version(actual.ReturnContext(), nil)
				assert.Nil(t, err)

				// Assert the returned version is a string matching a Semantic Version
				matched, _ := regexp.MatchString(versionPattern, version.Client.Version)
				assert.True(t, matched, "Version string should be a Semantic Version")
			} else {
				// Create instance of PodmanManager
				actual, err := NewPodmanManager(tc.config)

				// Assert an error is returned due to the bad unix socket
				// and that the connection returned is nil
				assert.Errorf(t, err, "Error: %s", "formatted")
				assert.Nil(t, actual)
			}
		})
	}
}

func TestCreateContainer_Integration(t *testing.T) {
	testCases := []struct {
		name           string
		containerImage string
		pass           bool
	}{
		{"real container", "docker.io/redhat/ubi9-minimal:latest", true},
		{"fake container", "example.com/libpoop/fake_alpine_nginx", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.pass {
				testContainerName := fmt.Sprintf("%x", rand.Int63())
				cm, err := NewContainerManager(&TestConfig.ContainerManagerConfig)
				if err != nil {
					t.Fatal(err)
				}

				_, _ = images.Pull(cm.ReturnContext(), tc.containerImage, nil)
				containerID, err := cm.CreateContainer(tc.containerImage, testContainerName)
				assert.Nil(t, err)

				inspectData, _ := containers.Inspect(cm.ReturnContext(), containerID, nil)

				assert.Equal(t, containerID, inspectData.ID, nil)

				// Remove test container
				force := true
				depend := true
				_, err = containers.Remove(cm.ReturnContext(), testContainerName, &containers.RemoveOptions{Force: &force, Depend: &depend})
				if err != nil {
					t.Fatal(err)
				}
			} else {
				randStr := fmt.Sprintf("%x", rand.Int63())
				cm, err := NewContainerManager(&TestConfig.ContainerManagerConfig)
				if err != nil {
					t.Fatal(err)
				}

				containerID, err := cm.CreateContainer(tc.containerImage, randStr)
				assert.Errorf(t, err, "Error: %s", "formatted")
				assert.Equal(t, "", containerID)
			}
		})
	}
}

func TestStartContainer_Integration(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		containerImage string
		pass           bool
	}{
		{"real container", "docker.io/redhat/ubi9-minimal:latest", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.pass {
				testContainerName := fmt.Sprintf("%x", rand.Int63())
				cm, err := NewContainerManager(&TestConfig.ContainerManagerConfig)
				if err != nil {
					t.Fatal(err)
				}

				_, _ = images.Pull(cm.ReturnContext(), tc.containerImage, nil)
				containerID, err := cm.CreateContainer(tc.containerImage, testContainerName)
				assert.Nil(t, err)

				// Start container
				err = cm.StartContainer(testContainerName)
				if err != nil {
					t.Fatal(err)
				}

				// Check container started
				inspectData, _ := containers.Inspect(cm.ReturnContext(), containerID, nil)
				assert.True(t, inspectData.State.Running)

				// Remove test container
				force := true
				depend := true
				_, err = containers.Remove(cm.ReturnContext(), testContainerName, &containers.RemoveOptions{Force: &force, Depend: &depend})
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestStopContainer_Integration(t *testing.T) {
	testCases := []struct {
		name           string
		containerImage string
		pass           bool
	}{
		{"real container", "docker.io/redhat/ubi9-minimal:latest", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.pass {
				testContainerName := fmt.Sprintf("%x", rand.Int63())
				cm, err := NewContainerManager(&TestConfig.ContainerManagerConfig)
				if err != nil {
					t.Fatal(err)
				}

				_, _ = images.Pull(cm.ReturnContext(), tc.containerImage, nil)
				containerID, err := cm.CreateContainer(tc.containerImage, testContainerName)
				assert.Nil(t, err)

				// Start container
				err = cm.StartContainer(testContainerName)
				if err != nil {
					t.Fatal(err)
				}

				// Check container started
				inspectData, _ := containers.Inspect(cm.ReturnContext(), containerID, nil)
				assert.True(t, inspectData.State.Running)

				// Stop Container
				err = cm.StopContainer(testContainerName)
				if err != nil {
					t.Fatal(err)
				}

				// Check container stopped
				inspectData, _ = containers.Inspect(cm.ReturnContext(), containerID, nil)
				assert.False(t, inspectData.State.Running)

				// Remove test container
				force := true
				depend := true
				_, err = containers.Remove(cm.ReturnContext(), testContainerName, &containers.RemoveOptions{Force: &force, Depend: &depend})
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestRemoveContainer_Integration(t *testing.T) {
	testCases := []struct {
		name           string
		containerImage string
		pass           bool
	}{
		{"real container", "docker.io/redhat/ubi9-minimal:latest", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.pass {
				testContainerName := fmt.Sprintf("%x", rand.Int63())
				cm, err := NewContainerManager(&TestConfig.ContainerManagerConfig)
				if err != nil {
					t.Fatal(err)
				}

				_, _ = images.Pull(cm.ReturnContext(), tc.containerImage, nil)

				s := specgen.NewSpecGenerator(tc.containerImage, false)
				s.Name = testContainerName
				s.Command = []string{"sleep", "infinity"}

				containerResponse, err := containers.CreateWithSpec(cm.ReturnContext(), s, nil)
				assert.Nil(t, err)

				// Start container
				err = containers.Start(cm.ReturnContext(), testContainerName, nil)
				if err != nil {
					t.Fatal(err)
				}

				// Check container started
				inspectData, _ := containers.Inspect(cm.ReturnContext(), containerResponse.ID, nil)
				assert.True(t, inspectData.State.Running)

				// Remove test container
				err = cm.RemoveContainer(testContainerName)
				if err != nil {
					t.Fatal(err)
				}

				// Check it's deleted
				_, err = containers.Inspect(cm.ReturnContext(), containerResponse.ID, nil)
				assert.Contains(t, err.Error(), "no container with name or ID")
			}
		})
	}
}

func TestPullImage_Integration(t *testing.T) {
	testCases := []struct {
		name           string
		containerImage string
		pass           bool
	}{
		{"real container", "docker.io/redhat/ubi9-minimal:latest", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.pass {
				cm, err := NewContainerManager(&TestConfig.ContainerManagerConfig)
				if err != nil {
					t.Fatal(err)
				}

				// Try pull image
				err = cm.PullImage(tc.containerImage)

				// Insure error is nil
				assert.Nil(t, err)

				// Verify image pulled
				imageExists, err := images.Exists(cm.ReturnContext(), tc.containerImage, nil)
				assert.Nil(t, err)
				assert.True(t, imageExists)
			}
		})
	}
}

func TestListContainer_Integration(t *testing.T) {
	testCases := []struct {
		name           string
		containerImage string
		pass           bool
	}{
		{"real container", "docker.io/redhat/ubi9-minimal:latest", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.pass {
				testContainerName := fmt.Sprintf("%x", rand.Int63())
				cm, err := NewContainerManager(&TestConfig.ContainerManagerConfig)
				if err != nil {
					t.Fatal(err)
				}

				_, _ = images.Pull(cm.ReturnContext(), tc.containerImage, nil)

				s := specgen.NewSpecGenerator(tc.containerImage, false)
				s.Labels = map[string]string{
					"pulpodControlled": "true",
				}
				s.Name = testContainerName
				s.Command = []string{"sleep", "infinity"}

				_, err = containers.CreateWithSpec(cm.ReturnContext(), s, nil)
				assert.Nil(t, err)

				// Start container
				err = containers.Start(cm.ReturnContext(), testContainerName, nil)
				if err != nil {
					t.Fatal(err)
				}

				// Check containerList returns our created container
				containerList, err := cm.List()
				if err != nil {
					t.Fatal(err)
				}
				println(containerList)
				assert.Contains(t, containerList, testContainerName)

				// Remove test container
				err = cm.RemoveContainer(testContainerName)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				testContainerName := fmt.Sprintf("%x", rand.Int63())
				cm, err := NewContainerManager(&TestConfig.ContainerManagerConfig)
				if err != nil {
					t.Fatal(err)
				}

				_, _ = images.Pull(cm.ReturnContext(), tc.containerImage, nil)

				s := specgen.NewSpecGenerator(tc.containerImage, false)
				s.Name = testContainerName
				s.Command = []string{"sleep", "infinity"}

				_, err = containers.CreateWithSpec(cm.ReturnContext(), s, nil)
				assert.Nil(t, err)

				// Start container
				err = containers.Start(cm.ReturnContext(), testContainerName, nil)
				if err != nil {
					t.Fatal(err)
				}

				// Check containerList returns our created container
				containerList, err := cm.List()
				if err != nil {
					t.Fatal(err)
				}

				assert.NotContains(t, containerList, testContainerName)

				// Remove test container
				err = cm.RemoveContainer(testContainerName)
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
