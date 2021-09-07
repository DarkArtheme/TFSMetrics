package azure

import (
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/tfvc"
)

type AzureInterface interface {
	Connect(config config)              // Подключение к Azure DevOps
	TfvcClientConnection(config) error  // для Repository.Open()
	ListOfProjects() ([]*string, error) // Получаем список проектов

	GetChangesets() ([]*int, error)                                      // Получает все id ченджсетов проекта
	GetChangesetChanges(id *int) (*ChangeSet, error)                     // получает все изминения для конкретного changeSet
	ChangedRows(currentFielUrl string, PreviusFileUrl string) (int, int) // Принимает ссылки на разные версии файлов возвращает Добавленные и Удаленные строки
}

type ChangeSet struct {
	ProjectName string
	Id          int
	Author      string
	Email       string
	AddedRows   int
	DeletedRows int
	Date        time.Time
	Message     string
	Hash        string
}

type Azure struct {
	Config     *config
	Connection *azuredevops.Connection
	TfvcClient tfvc.Client
}

func NewAzure(config *config) *Azure {
	return &Azure{
		Config: config,
	}
}

func (a *Azure) Connect() {
	organizationUrl := a.Config.OrganizationUrl
	personalAccessToken := a.Config.Token

	connection := azuredevops.NewPatConnection(organizationUrl, personalAccessToken)
	a.Connection = connection
}

func (a *Azure) TfvcClientConnection() error {
	tfvcClient, err := tfvc.NewClient(a.Config.Context, a.Connection)
	if err != nil {
		return err
	}
	a.TfvcClient = tfvcClient
	return nil
}

func (a *Azure) ListOfProjects() ([]*string, error) {
	coreClient, err := core.NewClient(a.Config.Context, a.Connection)
	if err != nil {
		return nil, err
	}

	resp, err := coreClient.GetProjects(a.Config.Context, core.GetProjectsArgs{})
	if err != nil {
		return nil, err
	}
	projectNames := []*string{}
	for _, project := range resp.Value {
		projectNames = append(projectNames, project.Name)
	}
	return projectNames, nil
}

func (a *Azure) GetChangesets(nameOfProject string) ([]*int, error) {
	changeSets, err := a.TfvcClient.GetChangesets(a.Config.Context, tfvc.GetChangesetsArgs{Project: &nameOfProject})
	if err != nil {
		return nil, err
	}
	changeSetIDs := []*int{}
	for _, v := range *changeSets {
		changeSetIDs = append(changeSetIDs, v.ChangesetId)
	}
	return changeSetIDs, nil
}

func (a *Azure) GetChangesetChanges(id *int, project string) (*ChangeSet, error) {
	// changesHash, err := a.TfvcClient.GetChangesetChanges(a.config.Context, tfvc.GetChangesetChangesArgs{Id: id})
	// if err != nil {
	// 	return &tfsmetrics.Commit{}, err
	// }
	changes, err := a.TfvcClient.GetChangeset(a.Config.Context, tfvc.GetChangesetArgs{Id: id, Project: &project})
	if err != nil {
		return &ChangeSet{}, err
	}
	messg := ""
	if changes.Comment != nil {
		messg = *changes.Comment
	}

	commit := &ChangeSet{
		ProjectName: project,
		Id:          *id,
		Author:      *changes.Author.DisplayName,
		Email:       *changes.Author.UniqueName,
		Date:        changes.CreatedDate.Time,
		Message:     messg,
	}
	// fmt.Println(changesHash.Value[0].Item["hashValue"])
	return commit, nil
}
