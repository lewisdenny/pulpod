package containermanager

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/bindings/system"
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
		versionPattern string = `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.
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
		{"real container", "quay.io/libpod/alpine_nginx", true},
		{"fake container", "example.com/libpoop/fake_alpine_nginx", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.pass {
				rand.Seed(uint64(time.Now().UnixNano()))
				randStr := fmt.Sprintf("%x", rand.Int63())
				cm, err := NewContainerManager(&TestConfig.ContainerManagerConfig)
				if err != nil {
					t.Fatal(err)
				}

				_, _ = images.Pull(cm.ReturnContext(), tc.containerImage, nil)
				containerID, err := cm.CreateContainer(tc.containerImage, randStr)
				assert.Nil(t, err)
				inspectData, err := containers.Inspect(cm.ReturnContext(), containerID, nil)
				assert.Equal(t, containerID, inspectData.ID, nil)
			} else {
				rand.Seed(uint64(time.Now().UnixNano()))
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