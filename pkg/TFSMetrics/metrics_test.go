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

	projects, err := azure.ListOfProjects()
	require.NoError(t, err)

	// var wg sync.WaitGroup
	// commitChan := make(chan Commit)

	// go func() {
	// 	wg.Add(1)
	// 	a := []Commit{}
	// 	for {
	// 		c := <-commitChan
	// 		a = append(a, c)
	// 		fmt.Println(c)
	// 	}
	// }()

	for _, project := range projects {
		commmits := &commitsCollection{
			nameOfProject: *project,
			azure:         azure,
		}
		err := commmits.Open()
		require.NoError(t, err)
		iter, err := commmits.GetCommitIterator()
		require.NoError(t, err)

		for commit, err := iter.Next(); err == nil; commit, err = iter.Next() {
			fmt.Println(commit)
		}

		// stopChan := make(chan struct{})
		// stop := false
		// wg.Add(1)
		// go func() {
		// 	defer wg.Done()
		// 	for {
		// 		select {
		// 		case <-stopChan:
		// 			return
		// 		default:
		// 			go func(stop *bool) {
		// 				commit, err := iter.Next()
		// 				if err != nil {
		// 					if !*stop {
		// 						*stop = false
		// 						stopChan <- struct{}{}
		// 						return
		// 					}
		// 					return
		// 				}
		// 				commitChan <- *commit
		// 			}(&stop)
		// 		}
		// 	}
		// }()
		// wg.Wait()
	}

}
