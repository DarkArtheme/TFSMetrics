package cli_metrics

import (
	"go-marathon-team-3/pkg/tfsmetrics/azure"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfigFile(t *testing.T) {
	_, err := ReadConfigFile("../../../configs/config.yaml")
	assert.NoError(t, err)
}

func TestWriteConfigFile(t *testing.T) {
	config := azure.NewConfig()
	config.OrganizationUrl = "url.com"
	config.Token = "12345"
	err := WriteConfigFile("../../../configs/config.yaml", config)
	assert.NoError(t, err)
	readConfig, err := ReadConfigFile("../../../configs/config.yaml")
	assert.Equal(t, config.OrganizationUrl, readConfig.OrganizationUrl)
	assert.Equal(t, config.Token, readConfig.Token)
}
