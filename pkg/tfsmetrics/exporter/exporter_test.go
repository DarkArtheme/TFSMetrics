package exporter

import (
	"go-marathon-team-3/pkg/tfsmetrics"
	"go-marathon-team-3/pkg/tfsmetrics/azure"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_exporter_GetProjectMetrics(t *testing.T) {
	conf := azure.NewConfig()
	conf.OrganizationUrl = "https://dev.azure.com/GnivcTestTaskTeam3"
	conf.Token = "yem42urypxdzuhceovddboakqs7skiicinze2i2u2leqrvbgblcq"

	azure := azure.NewAzure(conf)
	azure.Connect()

	projects, err := azure.ListOfProjects()
	require.NoError(t, err)

	for _, project := range projects {
		commmits := tfsmetrics.NewCommitCollection(*project, azure, false, nil)
		err = commmits.Open()
		require.NoError(t, err)
		iter, err := commmits.GetCommitIterator()
		require.NoError(t, err)

		exp := NewExporter()
		exp.GetProjectMetrics(iter, *project)

	}
	wg := sync.WaitGroup{}
	serv := NewPrometheusServer(&wg, time.Second*5)
	serv.Start(":8080")
	time.Sleep(time.Second * 30)
	err = serv.Stop()
	assert.NoError(t, err)
	wg.Wait()
}
