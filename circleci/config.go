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

type taskConfig struct {
	TaskInfo struct {
		Storage struct {
			ProjectRoot string
		}
	}

	AuthenticationCerts struct {
		CaCert     []byte
		RunnerCert []byte
		RunnerKey  []byte
	}

	AWSRegion string
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

	taskConfig := taskConfig{}
	err = json.Unmarshal(bytes, &taskConfig)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	if !strings.HasPrefix(taskConfig.TaskInfo.Storage.ProjectRoot, "s3://") {
		return nil, fmt.Errorf("only S3 storage backend is supported")
	}

	return &StorageConfig{
		ServiceURL:     serviceUri(),
		AuthCA:         taskConfig.AuthenticationCerts.CaCert,
		AuthClientCert: taskConfig.AuthenticationCerts.RunnerCert,
		AuthClientKey:  taskConfig.AuthenticationCerts.RunnerKey,
		AWSRegion:      taskConfig.AWSRegion,

		StorageRoot: taskConfig.TaskInfo.Storage.ProjectRoot,
	}, nil
}
