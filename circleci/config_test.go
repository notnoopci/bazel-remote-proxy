package circleci

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCircleCIConfig_Legacy(t *testing.T) {
	config := `
{
	"TaskInfo": {
		"Storage": {
			"ProjectRoot": "s3://example/root"
		}
	},
	"AWSRegion": "us-test-1"
}`

	c, err := parseCirclecIConfig([]byte(config))
	assert.NoError(t, err)
	assert.Equal(t, "s3://example/root", c.StorageRoot)
	assert.Equal(t, "us-test-1", c.AWSRegion)

}

func TestParseCircleCIConfig_New(t *testing.T) {
	config := `
{
	"Dispatched": {
		"TaskInfo": {
			"Storage": {
				"ProjectRoot": "s3://example/root"
			}
		},
		"AWSRegion": "us-test-1"
	}
}`

	c, err := parseCirclecIConfig([]byte(config))
	assert.NoError(t, err)
	assert.Equal(t, "s3://example/root", c.StorageRoot)
	assert.Equal(t, "us-test-1", c.AWSRegion)

}
