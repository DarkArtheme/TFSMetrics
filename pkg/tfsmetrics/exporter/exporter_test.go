package exporter

import (
	"fmt"
	"go-marathon-team-3/pkg/tfsmetrics"
	"go-marathon-team-3/pkg/tfsmetrics/azure"
	"testing"

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

	met := make(map[string]*ByAuthor)
	exp := NewExporter()
	for _, project := range projects {
		fmt.Println(*project)
		commmits := tfsmetrics.NewCommitCollection(*project, azure, false, nil)
		err = commmits.Open()
		require.NoError(t, err)
		iter, err := commmits.GetCommitIterator()
		require.NoError(t, err)

		met = exp.GetDataByAuthor(iter, "Андрей Назаренко", *project)

	}
	for k, v := range met {
		fmt.Println(k, v)
	}
	// wg := sync.WaitGroup{}
	// serv := NewPrometheusServer(&wg, time.Second*5)
	// serv.Start(":8080")
	// wg.Wait()
}
