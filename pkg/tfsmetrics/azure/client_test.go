package azure

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAzure_GetChangesetChanges(t *testing.T) {
	conf := NewConfig()
	conf.OrganizationUrl = "https://dev.azure.com/GnivcTestTaskTeam3"
	conf.Token = "yem42urypxdzuhceovddboakqs7skiicinze2i2u2leqrvbgblcq"

	azure := NewAzure(conf)
	azure.Connect()
	azure.TfvcClientConnection()
	projects, err := azure.ListOfProjects()
	assert.NoError(t, err)

	for _, p := range projects {
		changesets, _ := azure.GetChangesets(*p)
		for _, v := range changesets {
			fmt.Println(azure.GetChangesetChanges(v, *p))
		}
	}

}
