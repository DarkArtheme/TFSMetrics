package tfsmetrics

import (
	"fmt"
	"go-marathon-team-3/pkg/TFSMetrics/azure"
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
<<<<<<< HEAD
	azure.TfvcClientConnection()
	projects, _ := azure.ListOfProjects()

	for _, project := range projects {
		commmits := NewCommitCollection(*project, azure)
		iter, err := commmits.GetCommitIterator()
		require.NoError(t, err)
		for commit, err := iter.Next(); err == nil; commit, err = iter.Next() {
			fmt.Println(commit)
		}
=======

	projects, err := azure.ListOfProjects()
	require.NoError(t, err)

	project := projects[1]
	store, err := TestStore()
	require.NoError(t, err)
	defer store.Close()
	defer func() {
		os.Remove(store.db.Path())
	}()
	// for _, project := range projects {
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
	cacher := NewCacher(*project, store)
	newIter, err := cacher.Cache(iter)
	require.NoError(t, err)

	for commit, err := newIter.Next(); err == nil; commit, err = newIter.Next() {
		fmt.Println(commit)
>>>>>>> cli-cache
	}
	// }

}
