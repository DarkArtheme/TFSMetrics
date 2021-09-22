package azure

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
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

func NewAzure(config *Config) *Azure {
	return &Azure{
		Config: config,
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

func (a *Azure) ChangedRows(currentFileUrl string, PreviousFileUrl string) (int, int, error) {

	//TODO: РАЗБИТЬ МЕТОД НА НЕСКОЛЬКО МАЛЕНЬКИХ

	//1) СКАЧИВАНИЕ ФАЙЛОВ
	filepath1 := ""
	filepath2 := ""

	out1, err := os.Create(filepath1) //создание нового файла для currentFielUrl
	if err != nil {
		return 0, 0, err
	}
	defer out1.Close()

	out2, err := os.Create(filepath2) //создание нового файла для PreviusFileUrl
	if err != nil {
		return 0, 0, err
	}
	defer out2.Close()

	resp1, err := http.Get(currentFileUrl) //получаем актуальный файл
	if err != nil {
		return 0, 0, err
	}
	defer resp1.Body.Close()

	resp2, err := http.Get(PreviousFileUrl) //получаем предыдущий файл
	if err != nil {
		return 0, 0, err
	}
	defer resp2.Body.Close()

	_, err = io.Copy(out1, resp1.Body) //файл для currentFiel записан
	if err != nil {
		return 0, 0, err
	}

	_, err = io.Copy(out2, resp2.Body) //файл PreviusFile записан
	if err != nil {
		return 0, 0, err
	}

	//2) ОТКРЫТИЕ ФАЙЛОВ

	//CurrentFileData, err := ioutil.ReadFile(filepath1)
	//if err != nil {
	//	return 0, 0, err
	//}

	//PreviusFileData, err := ioutil.ReadFile(filepath2)
	//if err != nil {
	//	return 0, 0, err
	//}

	//3) ОПРЕДЕЛЕНИЕ КОЛЛИЧЕСТВА СТРОК
	savedRows := 0
	deletedRows := 0
	allRows := 0

	//Считать хэши строк или напрямую сравнивать строкуи из CurrentFileData и PreviusFileData??????

	addedRows := allRows - savedRows

	return addedRows, deletedRows, err
}

// заглушка чтобы избавиться от ошибки нереализованного интерфейса
func (a *Azure) GetItemVersions(ChangesUrl string) (int, int) {
	return 0, 0
}
