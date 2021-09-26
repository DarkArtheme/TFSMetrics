package azure

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/tfvc"
)

type AzureInterface interface {
	Azure() *Azure
	Connect()                           // Подключение к Azure DevOps
	TfvcClientConnection() error        // для Repository.Open()
	ListOfProjects() ([]*string, error) // Получаем список проектов

	GetChangesets(nameOfProject string) ([]*int, error)                          // Получает все id ченджсетов проекта
	GetChangesetChanges(id *int, project string) (*ChangeSet, error)             // получает все изминения для конкретного changeSet
	GetItemVersions(ChangesUrl string) (int, int)                                // Находит искомую и предыдущую версию файла, возвращает их юрл'ы
	ChangedRows(currentFileUrl string, PreviousFileUrl string) (int, int, error) // Принимает ссылки на разные версии файлов возвращает Добавленные и Удаленные строки
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
	Config     *Config
	Connection *azuredevops.Connection
	TfvcClient tfvc.Client
}

func NewAzure(conf *Config) AzureInterface {
	return &Azure{
		Config: conf,
	}
}

func (a *Azure) Azure() *Azure {
	return a
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
	changes, err := a.TfvcClient.GetChangeset(a.Config.Context, tfvc.GetChangesetArgs{Id: id, Project: &project})
	if err != nil {
		return nil, err
	}
	messg := ""
	if changes.Comment != nil {
		messg = *changes.Comment
	}
	changesHash, err := a.TfvcClient.GetChangesetChanges(a.Config.Context, tfvc.GetChangesetChangesArgs{Id: id})
	if err != nil {
		return nil, err
	}
	//fmt.Println(changesHash.Value[0].Item.(map[string]interface{}))
	//fmt.Println(changesHash.Value[0].Item.(map[string]interface{})["path"].(string), changesHash.Value[0].Item.(map[string]interface{})["version"])

	version := fmt.Sprint(changesHash.Value[0].Item.(map[string]interface{})["version"]) //приводим к строке версию изменения
	//получаем кол-во добавленных и удаленных строк
	addedRows, deletedRows, err := a.ChangedRows(changesHash.Value[0].Item.(map[string]interface{})["path"].(string), version)

	commit := &ChangeSet{
		ProjectName: project,
		Id:          *id,
		Author:      *changes.Author.DisplayName,
		Email:       *changes.Author.UniqueName,
		Date:        changes.CreatedDate.Time,
		Message:     messg,
		AddedRows:   addedRows,
		DeletedRows: deletedRows,
	}
	//fmt.Println(changesHash.Value[0].Item.(map[string]interface{})["version"])
	return commit, nil
}

func (a *Azure) ChangedRows(currentFileUrl, version string) (int, int, error) {
	//1 GET FILES
	//get current version
	item, err := a.TfvcClient.GetItemContent(a.Config.Context, tfvc.GetItemContentArgs{Path: &currentFileUrl,
		VersionDescriptor: &git.TfvcVersionDescriptor{Version: &version}})
	if err != nil {
		return 0, 0, err
	}
	b1, err := io.ReadAll(item)
	if err != nil { //Read All File in array byte
		return 0, 0, err
	}

	//get previous version
	item1, err := a.TfvcClient.GetItemContent(a.Config.Context, tfvc.GetItemContentArgs{Path: &currentFileUrl,
		VersionDescriptor: &git.TfvcVersionDescriptor{Version: &version, VersionOption: &git.TfvcVersionOptionValues.Previous}})
	if err != nil {
		return a.getAddedRowsOneFile(&b1) //если нет прошлой версии считаем кол-во строк в текущем файле
	}
	b2, err := io.ReadAll(item1)
	if err != nil {
		return 0, 0, err
	}

	//считаем добаленные и удаленные строки
	addedRows, deletedRows := Diff(string(b2), string(b1))
	return addedRows, deletedRows, nil
}

func (a *Azure) getAddedRowsOneFile(fileInBytes *[]byte) (int, int, error) {
	transformString := string(*fileInBytes)                     //transform array byte in string
	arrTransformStrings := strings.Split(transformString, "\n") //split string
	return len(arrTransformStrings), 0, nil
}

// заглушка чтобы избавиться от ошибки нереализованного интерфейса
func (a *Azure) GetItemVersions(ChangesUrl string) (int, int) {
	changesets, _ := a.GetChangesets(ChangesUrl)
	if len(changesets) > 1 {
		return *(changesets)[0], *(changesets)[1]
	}

	return *(changesets)[0], 0
}
