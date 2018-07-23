package circleci

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

type StorageConfig struct {
	ServiceURL string

	AuthCA         []byte
	AuthClientCert []byte
	AuthClientKey  []byte

	StorageRoot string
	AWSRegion   string
}

type dispatchedConfig struct {
	TaskInfo struct {
		Storage struct {
			ProjectRoot string
		}
	}
	AWSRegion string
}

type taskConfig struct {
	// legacy taskconfig had dispatch config inlined
	// so embedding it here
	dispatchedConfig

	// new taskConfig uses explicit Dispatched field
	Dispatched dispatchedConfig

	AuthenticationCerts struct {
		CaCert     []byte
		RunnerCert []byte
		RunnerKey  []byte
	}
}

func (c *taskConfig) projectRoot() string {
	r := c.Dispatched.TaskInfo.Storage.ProjectRoot
	if r == "" {
		r = c.TaskInfo.Storage.ProjectRoot
	}
	return r
}

func (c *taskConfig) awsRegion() string {
	r := c.Dispatched.AWSRegion
	if r == "" {
		r = c.AWSRegion
	}
	return r
}

// find the build agent service
func serviceUri() string {
	if ips, _ := net.LookupIP("circleci-internal-outer-build-agent"); len(ips) > 0 {
		return "https://circleci-internal-outer-build-agent:5500"
	}

	return "https://localhost:5500"
}

func ReadCircleCIConfig() (*StorageConfig, error) {
	configPath := os.Getenv("CIRCLE_INTERNAL_CONFIG")
	if configPath == "" {
		return nil, fmt.Errorf("running outside a circleci build")
	}

	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}

	return parseCirclecIConfig(bytes)
}

func parseCirclecIConfig(bytes []byte) (*StorageConfig, error) {
	taskConfig := taskConfig{}
	err := json.Unmarshal(bytes, &taskConfig)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	projectRoot := taskConfig.projectRoot()
	if !strings.HasPrefix(projectRoot, "s3://") {
		return nil, fmt.Errorf("only S3 storage backend is supported")
	}

	return &StorageConfig{
		ServiceURL:     serviceUri(),
		AuthCA:         taskConfig.AuthenticationCerts.CaCert,
		AuthClientCert: taskConfig.AuthenticationCerts.RunnerCert,
		AuthClientKey:  taskConfig.AuthenticationCerts.RunnerKey,
		AWSRegion:      taskConfig.awsRegion(),
		StorageRoot:    projectRoot,
	}, nil
}
