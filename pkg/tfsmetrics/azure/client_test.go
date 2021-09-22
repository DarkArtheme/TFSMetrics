package azure

import (
	"errors"
	"go-marathon-team-3/pkg/tfsmetrics/mock_tfvc_test"
	"time"

	"testing"

	"github.com/golang/mock/gomock"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/tfvc"
	"github.com/microsoft/azure-devops-go-api/azuredevops/webapi"
	"github.com/stretchr/testify/assert"
)

func TestAzure_GetChangesetChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedClient := mock_tfvc_test.NewMockClient(ctrl)

	conf := NewConfig()
	azure := Azure{
		Config:     conf,
		TfvcClient: mockedClient,
	}
	cs := ChangeSet{
		ProjectName: "project",
		Id:          1,
		Author:      "Ivan",
		Email:       "example@example.com",
		AddedRows:   0,
		DeletedRows: 0,
		Date:        time.Now(),
		Message:     "hello world",
		Hash:        "",
	}

	// правильная работа, без ощибки
	mockedClient.
		EXPECT().
		GetChangeset(azure.Config.Context, tfvc.GetChangesetArgs{Id: &cs.Id, Project: &cs.ProjectName}).
		Return(&git.TfvcChangeset{
			Author:      &webapi.IdentityRef{DisplayName: &cs.Author, UniqueName: &cs.Email},
			CreatedDate: &azuredevops.Time{Time: cs.Date},
			Comment:     &cs.Message,
		}, nil)

	changeSet, err := azure.GetChangesetChanges(&cs.Id, cs.ProjectName)
	assert.NoError(t, err)
	assert.Equal(t, &cs, changeSet)

	// azure возвращает ошибку
	cs.Id += 2
	mockedClient.
		EXPECT().
		GetChangeset(azure.Config.Context, tfvc.GetChangesetArgs{Id: &cs.Id, Project: &cs.ProjectName}).
		Return(nil, errors.New("error"))

	changeSet, err = azure.GetChangesetChanges(&cs.Id, cs.ProjectName)
	assert.Error(t, err)
	assert.Nil(t, changeSet)
}
