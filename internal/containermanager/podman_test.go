package containermanager

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/containers/podman/v5/pkg/bindings/system"
	"github.com/openstack-tooling/pulpod/internal/config"
	"github.com/openstack-tooling/pulpod/logging"
	"github.com/stretchr/testify/assert"
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

	type testCase struct {
		config *config.ContainerManagerConfig
		pass   bool
	}

	t.Run("valid config", func(t *testing.T) {
		tests := []testCase{
			{config: &TestConfig.ContainerManagerConfig, pass: true},
			{config: &invalidConfig, pass: false},
		}

		for _, test := range tests {
			if test.pass {
				// Create instance of PodmanManager
				actual, err := NewPodmanManager(test.config)

				// Assert there is no error and connection is not nil
				assert.NoError(t, err)
				assert.NotNil(t, actual.ReturnContext)

				// Fetch the version of the Podman server to test connection
				// and unsure no error is returned
				version, err := system.Version(actual.ReturnContext(), nil)
				assert.Nil(t, err)

				// Assert the returned version is a string matching a Semantic Version
				matched, _ := regexp.MatchString(versionPattern, version.Client.Version)
				assert.True(t, matched, "Version string should be a Semantic Version")
			} else {
				// Create instance of PodmanManager
				actual, err := NewPodmanManager(test.config)

				// Assert an error is returned due to the bad unix socket
				// and that the connection returned is nil
				assert.Errorf(t, err, "Error: %s", "formatted")
				assert.Nil(t, actual)
			}
		}
	})
}
