package exporter

import (
	"go-marathon-team-3/pkg/tfsmetrics"
	"go-marathon-team-3/pkg/tfsmetrics/azure"
	"go-marathon-team-3/pkg/tfsmetrics/store"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	store, err := store.TestStore()
	require.NoError(t, err)
	defer store.Close()
	defer func() {
		os.Remove(store.DB.Path())
	}()

	project := projects[1]
	commmits := tfsmetrics.NewCommitCollection(*project, azure, true, store)
	err = commmits.Open()
	require.NoError(t, err)
	iter, err := commmits.GetCommitIterator()
	require.NoError(t, err)

	exp := NewExporter()
	exp.GetProjectMetrics(iter)

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
