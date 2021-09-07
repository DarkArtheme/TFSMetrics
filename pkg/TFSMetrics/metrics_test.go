package tfsmetrics

import (
	"fmt"
	"go-marathon-team-3/pkg/TFSMetrics/azure"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_commitsCollection_Open(t *testing.T) {
	conf := azure.NewConfig()
	conf.OrganizationUrl = "https://dev.azure.com/GnivcTestTaskTeam3"
	conf.Token = "yem42urypxdzuhceovddboakqs7skiicinze2i2u2leqrvbgblcq"

	azure := azure.NewAzure(conf)
	azure.Connect()
	azure.TfvcClientConnection()
	projects, _ := azure.ListOfProjects()

	for _, project := range projects {
		commmits := &commitsCollection{
			nameOfProject: *project,
			azure:         azure,
		}
		iter, err := commmits.GetCommitIterator()
		require.NoError(t, err)
		for commit, err := iter.Next(); err == nil; commit, err = iter.Next() {
			fmt.Println(commit)
		}
	}
}
