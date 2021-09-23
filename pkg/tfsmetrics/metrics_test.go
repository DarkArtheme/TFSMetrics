package tfsmetrics

import (
	"errors"
	"go-marathon-team-3/pkg/tfsmetrics/azure"
	"go-marathon-team-3/pkg/tfsmetrics/mock"
	"go-marathon-team-3/pkg/tfsmetrics/repointerface"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/tfvc"
	"github.com/microsoft/azure-devops-go-api/azuredevops/webapi"
	"github.com/stretchr/testify/assert"
)

// func Test_commitsCollection_Open(t *testing.T) {
// 	conf := azure.NewConfig()
// 	conf.OrganizationUrl = "https://dev.azure.com/GnivcTestTaskTeam3"
// 	conf.Token = "yem42urypxdzuhceovddboakqs7skiicinze2i2u2leqrvbgblcq"

// 	azure := azure.NewAzure(conf)
// 	azure.Connect()

// 	projects, err := azure.ListOfProjects()
// 	require.NoError(t, err)

// 	store, err := store.TestStore()
// 	require.NoError(t, err)
// 	defer store.Close()
// 	defer func() {
// 		os.Remove(store.DB.Path())
// 	}()
// 	for _, project := range projects {
// 		commmits := NewCommitCollection(*project, azure, true, store)
// 		err := commmits.Open()
// 		require.NoError(t, err)
// 		iter, err := commmits.GetCommitIterator()
// 		require.NoError(t, err)

// 		for commit, err := iter.Next(); err == nil; commit, err = iter.Next() {
// 			fmt.Println(commit)
// 		}
// 	}

// }

func Test_iterator_Next_cahche_false(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedClient := mock.NewMockClient(ctrl)

	conf := azure.NewConfig()
	a := azure.Azure{
		Config:     conf,
		TfvcClient: mockedClient,
	}

	project := "project"

	c := repointerface.Commit{
		Id:          1,
		Author:      "Ivan",
		Email:       "example@example.com",
		AddedRows:   0,
		DeletedRows: 0,
		Date:        time.Now(),
		Message:     "hello world",
		Hash:        "",
	}

	iter := iterator{
		index:         0,
		commits:       []*int{&c.Id},
		nameOfProject: project,
		azure:         &a,
		cache:         false,
		store:         nil,
	}

	// правильная работа, без ощибки
	mockedClient.
		EXPECT().
		GetChangeset(a.Config.Context, tfvc.GetChangesetArgs{Id: &c.Id, Project: &project}).
		Return(&git.TfvcChangeset{
			Author:      &webapi.IdentityRef{DisplayName: &c.Author, UniqueName: &c.Email},
			CreatedDate: &azuredevops.Time{Time: c.Date},
			Comment:     &c.Message,
		}, nil)

	commit, err := iter.Next()
	assert.NoError(t, err)
	assert.Equal(t, &c, commit)

	// azure возвращает ошибку
	c.Id += 2
	iter.commits = append(iter.commits, &c.Id)
	mockedClient.
		EXPECT().
		GetChangeset(a.Config.Context, tfvc.GetChangesetArgs{Id: &c.Id, Project: &project}).
		Return(nil, errors.New("error"))

	commit, err = iter.Next()
	assert.Error(t, err)
	assert.Nil(t, commit)
}

func Test_iterator_Next_cahche_true(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedStore := mock.NewMockStore(ctrl)

	ctrlClient := gomock.NewController(t)
	defer ctrlClient.Finish()
	mockedClient := mock.NewMockClient(ctrlClient)

	conf := azure.NewConfig()
	a := azure.Azure{
		Config:     conf,
		TfvcClient: mockedClient,
	}

	project := "project"

	c := repointerface.Commit{
		Id:          1,
		Author:      "Ivan",
		Email:       "example@example.com",
		AddedRows:   0,
		DeletedRows: 0,
		Date:        time.Now(),
		Message:     "hello world",
		Hash:        "",
	}
	c2 := c
	c2.Id = 2

	iter := iterator{
		index:         0,
		commits:       []*int{&c.Id, &c2.Id},
		nameOfProject: project,
		azure:         &a,
		cache:         true,
		store:         mockedStore,
	}

	// не находит в бд берет из azure и записывает в базу
	mockedClient.
		EXPECT().
		GetChangeset(a.Config.Context, tfvc.GetChangesetArgs{Id: &c.Id, Project: &project}).
		Return(&git.TfvcChangeset{
			Author:      &webapi.IdentityRef{DisplayName: &c.Author, UniqueName: &c.Email},
			CreatedDate: &azuredevops.Time{Time: c.Date},
			Comment:     &c.Message,
		}, nil)

	mockedStore.
		EXPECT().
		FindOne(c.Id, project).
		Return(nil, errors.New("error"))

	mockedStore.
		EXPECT().
		Write(&c, project).
		Return(nil)

	commit, err := iter.Next()
	assert.NoError(t, err)
	assert.Equal(t, &c, commit)

	// берет из бд и возвращает
	mockedStore.
		EXPECT().
		FindOne(c2.Id, project).
		Return(&c2, nil)

	commit, err = iter.Next()
	assert.NoError(t, err)
	assert.Equal(t, &c2, commit)

}
