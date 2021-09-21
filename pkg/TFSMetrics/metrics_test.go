package tfsmetrics

import (
	"fmt"

	"go-marathon-team-3/pkg/tfsmetrics/azure"
	"go-marathon-team-3/pkg/tfsmetrics/cache"
	"go-marathon-team-3/pkg/tfsmetrics/store"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_commitsCollection_Open(t *testing.T) {
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
	for _, project := range projects {
		fmt.Println("start " + *project)
		commmits := &commitsCollection{
			nameOfProject: *project,
			azure:         azure,
		}
		err = commmits.Open()
		require.NoError(t, err)
		iter, err := commmits.GetCommitIterator()
		require.NoError(t, err)

		// for commit, err := iter.Next(); err == nil; commit, err = iter.Next() {
		// 	fmt.Println(commit)
		// }
		cacher := cache.NewCacher(*project, store)
		newIter, err := cacher.Cache(iter)
		require.NoError(t, err)

		for commit, err := newIter.Next(); err == nil; commit, err = newIter.Next() {
			fmt.Println(commit)
		}
	}

}
